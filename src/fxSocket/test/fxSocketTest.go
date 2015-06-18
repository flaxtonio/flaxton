package main

import (
	"io"
	"fxSocket"
	"golang.org/x/net/websocket"
	"os"
	"fmt"
)

// Echo the data received on the WebSocket.
func EchoServer(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

// This example demonstrates a trivial echo server.
func main() {
	switch os.Args[1] {
		case "server":
			{
				server := fxSocket.NewServer(":8888")
				server.OnConnection = func(s *fxSocket.Socket) {
					fmt.Println("connected", s.WS.RemoteAddr().String())
				}

				server.OnError = func(err error, s *fxSocket.Socket) {
					fmt.Println("Error !!", err.Error())
				}

				server.On("test", func(data []byte, s *fxSocket.Socket){
					fmt.Println(string(data))
					s.Emit("client_test", []byte("aaaaaaaaaaaaaaaa"))
				})
				server.Listen()
			}
		case "client":
			{
				client := fxSocket.NewClient("http://localhost:8888")
				client.OnConnection = func(s *fxSocket.Socket) {
					fmt.Println("connected", s.WS.RemoteAddr().String())
					s.Emit("test", []byte("!!!!!!!!!!!!!!!"))
				}

				client.OnError = func(err error, s *fxSocket.Socket) {
					fmt.Println("Error !!", err.Error())
				}

				client.On("client_test", func(data []byte, s *fxSocket.Socket){
					fmt.Println(string(data))
				})

				client.Connect()
			}
	}
}