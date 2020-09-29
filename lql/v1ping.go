package lql

import "github.com/gin-gonic/gin"

func v1Ping(c *gin.Context) (gin.H, error) {
	client, err := GinGetLqlClient(c)
	if err != nil {
		return nil, err
	}
	user := c.GetString("user")

	msg := `GET hosts
Columns: name`
	_, err = client.Request(c, msg, user, 1)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "ok"}, nil
}
