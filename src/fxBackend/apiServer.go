package fxBackend

import (
	"net/http"
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"sync"
)

type FlaxtonBackend struct {
	ListenHost string
}

var lock = sync.RWMutex{}

func (back *FlaxtonBackend) RunApiServer() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Post("/transfer/container", back.ContainerTransfer),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(back.ListenHost, api.MakeHandler()))
}

func (back *FlaxtonBackend) ContainerTransfer(w rest.ResponseWriter, r *rest.Request) {

}