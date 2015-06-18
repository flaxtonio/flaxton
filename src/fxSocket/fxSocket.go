package fxSocket

type SocketEvent struct {
	Name        	string 				`json:"name"`
	Handler     	func(interface{}) 	`json:"-"`
	HandlerData 	interface{}         `json:"handler_data"`
}

type Server struct {
	ListenAddress string `json:"listen_address"`
	Events []SocketEvent `json:"events"`
}

type Client struct {
	Events []SocketEvent `json:"events"`
}