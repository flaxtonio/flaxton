package main

import (
	"fxSocket"
	"os"
	"fmt"
)

func main() {
	switch os.Args[1] {
		case "parent":
			{
				s, err := fxSocket.NewParent(":8888")
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				s.On("test", func(data []byte, socket fxSocket.Socket){
					fmt.Println(string(data))
					socket.Emit("test_child", "Second Test client")
				})

				s.OnError = func(e error) {
					fmt.Println(e.Error())
				}

				s.Listen()
			}
		case "child":
			{
				c, err := fxSocket.NewChild("127.0.0.1:8888")
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
				c.On("test_child", func(data []byte, socket fxSocket.Socket){
					fmt.Println(string(data))
				})
				c.OnError = func(e error) {
					fmt.Println(e.Error())
				}
				c.Emit("test", "Test Message")
				c.WaitForEvents()
			}
	}
}