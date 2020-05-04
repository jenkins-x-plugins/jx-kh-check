package main

import (
	"github.com/jenkins-x/jx-logging/pkg/log"
)

// Entrypoint for the command
func main() {
	log.Logger().Infof("hey")
}
