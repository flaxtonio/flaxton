package lib

import (
	"github.com/Sirupsen/logrus"
)

// Central point for logging error messages, or error object
func LogError(message string, err error) {
	logrus.Error(message)
	if err != nil {
		logrus.Error(err.Error())
	}
}