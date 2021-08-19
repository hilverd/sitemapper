package main

import (
	"fmt"
	"net"
	"net/http"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:") // port number left empty to automatically choose one
	if err != nil {
		panic(err)
	}

	fmt.Printf("Serving on http://127.0.0.1:%v\n", listener.Addr().(*net.TCPAddr).Port)

	h := http.FileServer(http.Dir("./html"))
	panic(http.Serve(listener, h))
}
