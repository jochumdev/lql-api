package lql

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/webmeisterei/lql-api/internal/gncp"
	"github.com/webmeisterei/lql-api/internal/utils"

	log "github.com/sirupsen/logrus"
)

func connCreator(network string, address string) func() (net.Conn, error) {
	return func() (net.Conn, error) {
		return net.Dial(network, address)
	}
}

type Client struct {
	pool      *gncp.GncpPool
	logger    *log.Logger
	timeLimit int
}

func NewClient(minConn, maxConn int, network, address string) (*Client, error) {
	pool, err := gncp.NewPool(minConn, maxConn, connCreator(network, address))
	if err != nil {
		return nil, err
	}

	return &Client{
		pool:      pool,
		logger:    log.New(),
		timeLimit: 60,
	}, nil
}

func (c *Client) SetLogger(logger *log.Logger) {
	c.logger = logger
}

func (c *Client) Close() error {
	return c.pool.Close()
}

func (c *Client) modifyRaw(request, outputFormat, authUser string, limit int) (string, error) {
	request = strings.Replace(request, "\n\n", "\n", -1)
	request = strings.Trim(request, "\n")

	lines := strings.Split(request, "\n")
	if len(lines) == 0 {
		return "", errors.New("No newlines in the request")
	}

	isGet := false
	requestHeaderLine := -1
	hasFormat := false
	columnHeadersLine := -1
	keepAliveLine := -1
	hasAuthUser := false
	for n, line := range lines {
		if n == 0 {
			mt := strings.Split(line, " ")
			if strings.Trim(mt[0], " ") == "GET" {
				isGet = true
			}
			continue
		}

		hh := strings.Split(line, ":")
		h := strings.Trim(hh[0], " ")
		switch h {
		case "ResponseHeader":
			requestHeaderLine = n
			break
		case "OutputFormat":
			hasFormat = true
		case "ColumnHeaders":
			columnHeadersLine = n
		case "KeepAlive":
			keepAliveLine = n
		case "AuthUser":
			hasAuthUser = true
		default:
		}
	}

	if isGet {
		addedLines := 0
		if requestHeaderLine == -1 {
			lines = utils.StringArrayInsert(lines, 1, "ResponseHeader: fixed16")
			addedLines++
		} else {
			lines = utils.StringArrayReplace(lines, requestHeaderLine+addedLines, "ResponseHeader: fixed16")
		}
		if keepAliveLine == -1 {
			lines = utils.StringArrayInsert(lines, 1, "KeepAlive: on")
			addedLines++
		} else {
			lines = utils.StringArrayReplace(lines, keepAliveLine+addedLines, "KeepAlive: on")
		}

		if columnHeadersLine == -1 {
			lines = append(lines, "ColumnHeaders: on")
			addedLines++
		} else {
			lines = utils.StringArrayReplace(lines, columnHeadersLine+addedLines, "ColumnHeaders: on")
		}

		if !hasFormat && outputFormat != "" {
			lines = append(lines, fmt.Sprintf("OutputFormat: %s", outputFormat))
		}
		if !hasAuthUser && authUser != "" {
			lines = append(lines, fmt.Sprintf("AuthUser: %s", authUser))
		}
		if limit > 0 {
			lines = append(lines, fmt.Sprintf("Limit: %d", limit))
		}

	}

	// LQL requires two newlines as end of input
	return strings.Join(lines, "\n") + "\n\n", nil
}

func (c *Client) Request(context context.Context, request, authUser string, limit int) ([]gin.H, error) {
	rawResponse, err := c.RequestRaw(context, request, "json", authUser, limit)
	if err != nil {
		return nil, err
	}

	parsedJson := make([]interface{}, bytes.Count(rawResponse, []byte{'\n'}))
	json.Unmarshal(rawResponse, &parsedJson)

	headers := []string{}
	result := []gin.H{}
	for i, data := range parsedJson {
		if i == 0 {
			for _, header := range data.([]interface{}) {
				headers = append(headers, header.(string))
			}
			continue
		}

		myEntry := gin.H{}
		for n, header := range headers {
			myEntry[header] = data.([]interface{})[n]
		}

		result = append(result, myEntry)
	}

	return result, err
}

func (c *Client) RequestRaw(context context.Context, request, outputFormat, authUser string, limit int) ([]byte, error) {
	request, err := c.modifyRaw(request, outputFormat, authUser, limit)
	if err != nil {
		return nil, err
	}

	conn, err := c.pool.GetWithContext(context)
	if err != nil {
		return nil, err
	}

	c.logger.WithField("request", request).Debug("Writing request")
	_, err = conn.Write([]byte(request))
	if err != nil && !errors.Is(err, syscall.EPIPE) {
		c.logger.WithField("error", err).Debug("Removing failed connection")
		c.pool.Remove(conn)
		return nil, err
	} else if errors.Is(err, syscall.EPIPE) {
		c.pool.Remove(conn)

		// Destroy -> Create Connections until we don't get EPIPE.
		numTries := 0
		maxTries := c.pool.GetMaxConns() * 2
		for errors.Is(err, syscall.EPIPE) {
			c.logger.WithFields(log.Fields{"error": err, "num_tries": numTries, "max_tries": maxTries, "max_conns": c.pool.GetMaxConns()}).Debug("Trying to reconnect")

			conn, err = c.pool.GetWithContext(context)
			if err != nil {
				// Failed to get a connection, bailout
				return nil, err
			}

			_, err = conn.Write([]byte(request))
			if err != nil && !errors.Is(err, syscall.EPIPE) {
				// Other error than EPIPE, bailout
				c.logger.WithField("error", err).Debug("Removing failed connection")
				c.pool.Remove(conn)
				return nil, err
			} else if err == nil {
				// We are fine now
				break
			}

			c.pool.Remove(conn)

			numTries++
			if numTries >= maxTries {
				c.logger.WithField("error", err).Error("To much retries can't reconnect")
				// Bailout to much tries
				return nil, err
			}
		}
	}
	defer c.pool.Put(conn)

	tmpBuff := make([]byte, 1024)
	n, err := conn.Read(tmpBuff)
	if err != nil {
		return nil, err
	}

	idx := utils.BinarySearch(tmpBuff, '\n')
	if idx == -1 || idx < 15 {
		c.logger.WithField("output", string(tmpBuff[0:n])).Error("Empty output")
		return nil, errors.New("Empty output")
	}
	resultBuff := new(bytes.Buffer)
	resultBuff.Write(tmpBuff[idx+1 : n])

	line := string(tmpBuff[0:idx])
	if err != nil {
		return nil, err
	}
	statusCode, err := strconv.Atoi(line[0:3])
	if err != nil {
		return nil, err
	}
	length, err := strconv.Atoi(strings.Trim(line[5:15], " "))
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, err
	}

	for resultBuff.Len() < length {
		n, err := conn.Read(tmpBuff)
		if err != nil {
			return nil, err
		}
		resultBuff.Write(tmpBuff[0:n])
	}

	return resultBuff.Bytes(), nil
}
