package lql

import (
	"context"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Client interface {
	ClientCount() int
	IsAdmin(username string) bool
	SetLogger(logger *log.Logger)
	Close() error
	Request(context context.Context, request, authUser string, limit int) ([]gin.H, error)
	RequestRaw(context context.Context, request, outputFormat, authUser string, limit int) ([]byte, error)
}
