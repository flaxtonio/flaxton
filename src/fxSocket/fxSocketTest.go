package main

import (
	"io"
	"net/http"

	"golang.org/x/net/websocket"
	"os"
	"log"
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
				http.Handle("/echo", websocket.Handler(EchoServer))
				err := http.ListenAndServe(":12345", nil)
				if err != nil {
					panic("ListenAndServe: " + err.Error())
				}
			}
		case "client":
			{
				origin := "http://localhost/"
				url := "ws://localhost:12345/echo"
				ws, err := websocket.Dial(url, "", origin)
				if err != nil {
					log.Fatal(err)
				}
				for i:= 5; i < 15; i++ {
					if _, err := ws.Write([]byte(fmt.Sprintf("hello, world! %d\n", i))); err != nil {
						log.Fatal(err)
					}
					var msg = make([]byte, 512)
					var n int
					if n, err = ws.Read(msg); err != nil {
						log.Fatal(err)
					}
					fmt.Printf("Received: %s", msg[:n])
				}
			}
	}
}