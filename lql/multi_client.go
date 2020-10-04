package lql

import (
	"context"
	"io/ioutil"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
)

type MultiClient struct {
	minConn      int
	maxConn      int
	localSocket  string
	liveproxyDir string
	clients      map[string]Client
	usersWatcher *UsersWatcher
}

func NewMultiClient(minConn, maxConn int, localSocket, liveproxyDir string, multisiteUsersFile string) (Client, error) {
	uw, err := NewUsersWatcher(multisiteUsersFile)
	if err != nil {
		return nil, err
	}

	mc := &MultiClient{
		minConn:      minConn,
		maxConn:      maxConn,
		localSocket:  localSocket,
		liveproxyDir: liveproxyDir,
		clients:      make(map[string]Client),
		usersWatcher: uw,
	}

	err = mc.CreateClients()
	if err != nil {
		return nil, err
	}

	return mc, nil
}

func (c *MultiClient) ClientCount() int {
	return len(c.clients)
}

func (c *MultiClient) CreateClients() error {
	client, err := NewSingleClient(c.minConn, c.maxConn, "unix", c.localSocket)
	if err != nil {
		return err
	}
	c.clients["__local__"] = client

	files, err := ioutil.ReadDir(c.liveproxyDir)
	if err != nil {
		// Ignore listing errors and use the local client only
		return nil
	}

	var result error
	for _, file := range files {
		filePath := path.Join(c.liveproxyDir, file.Name())

		if file.Mode()&os.ModeSocket == 0 {
			log.Debugf("Ignoring non-socket file: %s", filePath)
			continue
		}

		client, err = NewSingleClient(c.minConn, c.maxConn, "unix", filePath)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		c.clients[file.Name()] = client
	}

	return result
}

func (c *MultiClient) IsAdmin(username string) bool {
	return c.usersWatcher.IsAdmin(username)
}

func (c *MultiClient) SetLogger(logger *log.Logger) {
	c.usersWatcher.SetLogger(logger)

	for _, client := range c.clients {
		client.SetLogger(logger)
	}
}

func (c *MultiClient) Close() (result error) {
	for _, client := range c.clients {
		if err := client.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (c *MultiClient) Request(context context.Context, request, authUser string, limit int) ([]gin.H, error) {
	result := []gin.H{}
	var resultErr error

	// Divide limit per client
	limits := []int{}
	perRequest := 0
	firstLimit := 0
	if limit > 0 {
		if limit > c.ClientCount() {
			perRequest = int(limit / c.ClientCount())
			firstLimit = perRequest + int(limit%c.ClientCount())
		} else {
			perRequest = 1
			firstLimit = 1
		}
	}
	for i := 0; i < len(c.clients); i++ {
		if i == 0 {
			limits = append(limits, firstLimit)
			continue
		}

		limits = append(limits, perRequest)
	}

	i := 0
	for _, client := range c.clients {
		tmpResult, err := client.Request(context, request, authUser, limits[i])
		if err != nil {
			resultErr = multierror.Append(resultErr, err)
			continue
		}

		if len(tmpResult) == 0 {
			continue
		}

		if len(result) > 0 {
			allFieldsStats := true
			for k := range result[0] {
				if len(k) > 6 && k[0:6] == "stats_" {
					allFieldsStats = true
				} else {
					allFieldsStats = false
				}
			}

			if allFieldsStats {
				for i, row := range tmpResult {
					for k, v := range row {
						result[i][k] = result[i][k].(float64) + v.(float64)
					}
				}
			} else {
				result = append(result, tmpResult...)
			}
		} else {
			result = append(result, tmpResult...)
		}

		i++

		// If we have limit < client count
		if limit > 0 && i > limit {
			break
		}
	}

	return result, resultErr
}

func (c *MultiClient) RequestRaw(context context.Context, request, outputFormat, authUser string, limit int) ([]byte, error) {
	result := []byte{}
	var resultErr error

	for _, client := range c.clients {
		tmpResult, err := client.RequestRaw(context, request, outputFormat, authUser, limit)
		if err != nil {
			resultErr = multierror.Append(resultErr, err)
			continue
		}

		result = append(result, tmpResult...)
	}

	return result, resultErr
}
