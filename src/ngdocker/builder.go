package ngdocker

import (
	"log"
)

func ExecuteArguments(args []string) {
	if args[1] != "ngdocker" {
		log.Panic("2nd Argument should be ngdocker, for building Nginx Docker Container")
		return
	}

	switch args[2] {

	}
}