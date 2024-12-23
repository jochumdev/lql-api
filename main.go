package main

import (
	"github.com/jochumdev/lql-api/cmd"
)

// @title LQL API
// @version 1.0
// @description This is the LQL API for your check_mk Server.

// @contact.name Developers
// @contact.url https://github.com/jochumdev/lql-api/issues
// @contact.email support@webmeisterei.com

// @license.name MIT
// @license.url https://github.com/jochumdev/lql-api/blob/master/LICENSE

// @BasePath /v1
func main() {
	cmd.Execute()
}
