package lql

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
)

const usersExporterFile = `from __future__ import print_function

import json

class MultiSiteUsers(object):
    def update(self, data):
        print(json.dumps(data));

multisite_users = MultiSiteUsers()

eval(open("%s").read())
`

type UserData struct {
	ForceAuthUserWebservice bool     `json:"force_authuser_webservice"`
	Looked                  bool     `json:"locked"`
	Roles                   []string `json:"roles"`
	ForceAuthUser           bool     `json:"force_authuser"`
	Alias                   string   `json:"alias"`
	StartUrl                string   `json:"start_url"`
}

type UsersWatcher struct {
	usersfile  string
	users      map[string]UserData
	lock       *sync.RWMutex
	logger     *log.Logger
	isWatching bool
	watcher    *fsnotify.Watcher
}

func NewUsersWatcher(usersfile string) (*UsersWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	uw := &UsersWatcher{
		usersfile:  usersfile,
		lock:       &sync.RWMutex{},
		isWatching: false,
		watcher:    watcher,
	}

	return uw, nil
}

func (uw *UsersWatcher) Close() {
	uw.watcher.Close()
}

func (uw *UsersWatcher) SetLogger(logger *log.Logger) {
	uw.logger = logger
}

func (uw *UsersWatcher) StartWatching() {
	go func() {
		for {
			select {
			case event, ok := <-uw.watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					uw.FetchUsers()
				}
			case err, ok := <-uw.watcher.Errors:
				if !ok {
					return
				}
				uw.logger.WithField("error", err).Error()
			}
		}
	}()

	uw.lock.Lock()
	uw.isWatching = true
	uw.lock.Unlock()

	err := uw.watcher.Add(uw.usersfile)
	if err != nil {
		uw.logger.WithField("error", err).Error()
	}
}

func (uw *UsersWatcher) IsAdmin(userName string) bool {
	uw.lock.RLock()
	if uw.users == nil {
		uw.lock.RUnlock()

		if !uw.isWatching {
			uw.StartWatching()
		}

		uw.FetchUsers()
		uw.lock.RLock()
	}
	defer uw.lock.RUnlock()

	userData, ok := uw.users[userName]
	if !ok {
		uw.logger.WithField("user_name", userName).Debug("Failed to fetch user from db")
		return false
	}

	for _, role := range userData.Roles {
		if role == "admin" {
			uw.logger.WithField("user_name", userName).Trace("User is admin")
			return true
		}
	}

	uw.logger.WithField("user_name", userName).Trace("User is not admin")
	return false
}

func (uw *UsersWatcher) FetchUsers() error {
	uw.logger.WithField("usersfile", uw.usersfile).Debug("Reading users")

	dir, err := ioutil.TempDir("", "lql-api")
	if err != nil {
		uw.logger.WithField("error", err).Error()
		return err
	}

	tmpfn := filepath.Join(dir, "lql-api-user-reader.py")
	if err := ioutil.WriteFile(tmpfn, []byte(fmt.Sprintf(usersExporterFile, uw.usersfile)), 0700); err != nil {
		uw.logger.WithField("error", err).Error()
		return err
	}

	cmd := exec.Command("python", tmpfn)
	uw.logger.WithField("args", cmd.Args).Debug("Executing")
	out, err := cmd.CombinedOutput()
	if err != nil {
		uw.logger.WithField("error", err).Error()
		return err
	}

	result := make(map[string]UserData, 1)
	if err = json.Unmarshal(out, &result); err != nil {
		uw.logger.WithField("error", err).Error()
		return err
	}

	uw.lock.Lock()
	uw.users = result
	uw.lock.Unlock()

	return nil
}
