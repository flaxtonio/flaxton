package fxSocket

import (
	"golang.org/x/net/websocket"
	"net/http"
	"io/ioutil"
	"strings"
	"io"
	"encoding/json"
	"errors"
)

type SocketMessage struct {
	Event 	string `json:"event"`
	Data 	[]byte `json:"data"`
}

type SocketComponent struct {
	Events 			map[string][]func([]byte, *Socket) 	`json:"-"`
	OnError			func(error, *Socket)         		`json:"-"`
	OnConnection	func(*Socket)       				`json:"-"`
	OnDisconnect	func(*Socket)       				`json:"-"`
}

type Socket struct {
	WS	*websocket.Conn		`json:"-"`
}

func (sock *SocketComponent) On(name string, handler func([]byte, *Socket)) {
	sock.Events[name] = append(sock.Events[name], handler)
}

func (sock *SocketComponent) WebSocketHandler(ws *websocket.Conn) {
	var (
		data 		[]byte
		err 		error
		sock_msg 	SocketMessage
		wss 		Socket
	)
	wss.WS = ws
	if sock.OnConnection != nil {
		sock.OnConnection(&wss)
	}

	for {
		data, err = ioutil.ReadAll(ws)
		if err != nil {
			// If there is no EOF then something went wrong
			if sock.OnError != nil && err != io.EOF {
				sock.OnError(err, &wss)
			}
			break
		}

		err = json.Unmarshal(data, &sock_msg)
		if err != nil {
			if sock.OnError != nil {
				sock.OnError(err, &wss)
			}
			continue
		}

		if _, ok := sock.Events[sock_msg.Event]; ok {
			for _, h := range sock.Events[sock_msg.Event] {
				h(sock_msg.Data,  &wss)
			}
		}
	}

	if sock.OnDisconnect != nil {
		sock.OnDisconnect(&wss)
	}
}

func (sock *Socket) Emit(event string, data []byte) error {
	if sock.WS == nil {
		return errors.New("There is no connection to emit for")
	}
	msg := SocketMessage{Event:event, Data:data}
	json_data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = sock.WS.Write(json_data)
	return err
}

type Server struct {
	SocketComponent
	ListenAddress 	string 			`json:"listen_address"`
	ServerMux 		*http.ServeMux 	`json:"-"`
}

func NewServer(listen_addr string) (srv Server) {
	srv.ServerMux = http.NewServeMux()
	srv.ListenAddress = listen_addr
	srv.Events = make(map[string][]func([]byte, *Socket))
	return
}


func (s *Server) Listen() (err error) {
	// Setting socket event handlers here
	s.ServerMux.Handle("/", websocket.Handler(s.WebSocketHandler))
	err = http.ListenAndServe(s.ListenAddress, s.ServerMux)
	return
}

type Client struct {
	SocketComponent
	ServerAddress 	string 			`json:"server_address"`
}

func NewClient(addr string) (client Client) {
	client.ServerAddress = addr
	client.Events = make(map[string][]func([]byte, *Socket))
	return
}

func (c *Client) Connect() {
	ws, err := websocket.Dial(strings.Replace(c.ServerAddress, "http://", "ws://", -1), "", c.ServerAddress)
	if err != nil {
		if c.OnError != nil {
			c.OnError(err, nil)
			return
		}
	}

	c.WebSocketHandler(ws)
}