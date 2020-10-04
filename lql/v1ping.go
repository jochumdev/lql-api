package lql

import "github.com/gin-gonic/gin"

func v1Ping(c *gin.Context) (gin.H, error) {
	client, err := GinGetLqlClient(c)
	if err != nil {
		return nil, err
	}
	user := c.GetString("user")
	if client.IsAdmin(user) {
		user = ""
	}

	msg := `GET hosts
Columns: name`
	_, err = client.Request(c, msg, user, client.ClientCount())
	if err != nil {
		Logger.Error(err)
	}

	return gin.H{"message": "pong"}, nil
}
