package fxBackend

import (
	"fxSocket"
)

var BackendService fxSocket.Parent

func Start(listen_address string) (err error) {
	BackendService, err = fxSocket.NewParent(listen_address)
	return
}

func RegisterEvents() {

}