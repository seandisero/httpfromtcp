package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
)

const inputFilePath = "messages.txt"

func getLinesChannel(f io.Reader) <-chan string {
	ch := make(chan string)

	go func(f io.Reader) {
		defer close(ch)

		b := make([]byte, 8, 8)
		currentLine := ""
		for {
			n, err := f.Read(b)
			if err != nil {
				if currentLine != "" {
					ch <- currentLine
					currentLine = ""
				}
				if errors.Is(err, io.EOF) {
					return
				}
				fmt.Printf("error: %s\n", err.Error())
				return
			}
			byteList := bytes.Split(b[:n], []byte("\n"))
			currentLine = currentLine + string(byteList[0])
			if len(byteList) > 1 {
				ch <- currentLine
				currentLine = string(byteList[1])
			}
		}
	}(f)

	return ch
}

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
		read := getLinesChannel(conn)
		for str := range read {
			fmt.Println(str)
		}
	}
}
