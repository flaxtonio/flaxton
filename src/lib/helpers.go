package lib

import (
	"os/exec"
	"log"
	"os"
)

func RandomString(n int) string {
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return string(out)
}
