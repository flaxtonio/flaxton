package fxdocker

import (
	"net/http"
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"sync"
	"strings"
)

type FxDaemonApi struct {
	Fxd *FxDaemon
}

var lock = sync.RWMutex{}

func (fx_api *FxDaemonApi) RunApiServer() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Post("/childs", fx_api.ChildServers),
		rest.Post("/child/iplookup", fx_api.ChildServerIpLookup),

		rest.Post("/transfer/container", fx_api.TransferContainer),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(fx_api.Fxd.ListenHost, api.MakeHandler()))
}

type IpLookupCall struct {
	IP string `json:"ip"`
	//TODO: Maybe here should be other fields for IP lookup functionality
}

// Lookup for getting visible IP address from child server daemon
func (fx_api *FxDaemonApi) ChildServerIpLookup(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(map[string]string{"ip": strings.Split(r.RemoteAddr,":")[0]})
}

type ChildrenCall struct {
	Type 	string 		`json:"type"`
	Hosts 	[]string 	`json:"hosts"`
}

func (fx_api *FxDaemonApi) ChildServers(w rest.ResponseWriter, r *rest.Request){
	children := ChildrenCall{}
	err := r.DecodeJsonPayload(&children)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wrong_type := false
	lock.Lock()
	switch children.Type {
		case "add":
			{
				for _, host := range children.Hosts  {
					fx_api.Fxd.AddChildServer(host)
				}
			}
		case "delete":
			{
				for _, host := range children.Hosts  {
					fx_api.Fxd.DeleteChildServer(host)
				}
			}
		default:
			{
				w.WriteJson(map[string]string{"status": "error", "message": "Type must be 'add' or 'delete' !"})
				wrong_type = true
			}
	}
	lock.Unlock()

	if !wrong_type {
		w.WriteJson(map[string]string{"status": "ok"})
	}
}

type TransferContainerCall struct {
	Name 		string          `json:"name"`
	Cmd 		string 			`json:"cmd"`
	ImageName 	string			`json:"image_name"`
	ImageId 	string          `json:"image_id"`
	NeedToRun	bool        	`json:"need_to_run"`
	Authorization string 		`json:"authorization"`
}

func (fx_api *FxDaemonApi) TransferContainer(w rest.ResponseWriter, r *rest.Request){
	transfer := TransferContainerCall{}
	err := r.DecodeJsonPayload(&transfer)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	container_id, err2 := fx_api.Fxd.TransferContainer(transfer)
	if err2 != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteJson(map[string]string{"container": container_id})
}