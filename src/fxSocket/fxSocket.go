package fxSocket

import (
	"golang.org/x/net/websocket"
	"net/http"
	"fmt"
)

type SocketEvent struct {
	Name        	string 					`json:"name"`
	Handler     	func(*websocket.Conn) 	`json:"-"`
	HandlerData 	interface{}         	`json:"handler_data"`
}

type Server struct {
	ListenAddress 	string 					`json:"listen_address"`
	/**
	Events with key -> value map , where key - event name , value - event handler function
	 */
	Events 			map[string]SocketEvent 	`json:"events"`
	ServerMux 		*http.ServeMux 			`json:"-"`
}

type Client struct {
	ServerAddress 	string 			`json:"server_address"`
	Events 			[]SocketEvent 	`json:"events"`
}

func NewServer(listen_addr string) (srv Server) {
	srv.ServerMux = http.NewServeMux()
	srv.ListenAddress = listen_addr
	return
}

func (s *Server) Listen() (err error) {
	// Setting socket event handlers here
	for _, ev := range s.Events {
		s.ServerMux.Handle(fmt.Sprintf("/%s", ev.Name), websocket.Handler(ev))
	}
	err = http.ListenAndServe(s.ListenAddress, s.ServerMux)
	return
}

func (s *Server) AddEvent() {

}