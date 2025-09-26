package main

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/seandisero/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		slog.Error("error creating listener", "error", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("connection not accepted", "error", err)
			return
		}
		slog.Info("connection accepted")
		req, err := request.RequestFromReader(conn)
		if err != nil {
			slog.Error("could nto read request from connection")
			return
		}
		fmt.Printf(
			"Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
			req.RequestLine.Method,
			req.RequestLine.RequestTarget,
			req.RequestLine.HttpVersion,
		)
		fmt.Println("Headers:")
		for key, val := range req.Headers {
			fmt.Printf("- %s: %s\n", key, val)

		}
		fmt.Println("Body: ")
		fmt.Println(string(req.Body))
	}
}
