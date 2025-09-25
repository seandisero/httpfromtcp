package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"os"
)

func main() {
	udpaddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		slog.Error("could not resolve udp addr", "error", err)
		return
	}

	udpconn, err := net.DialUDP("udp", nil, udpaddr)
	if err != nil {
		slog.Error("could not make udp connection", "error", err)
		return
	}
	defer udpconn.Close()

	buffer := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf(">")
		str, err := buffer.ReadString('\n')
		if err != nil {
			slog.Error("could not read string from buffer", "error", err)
			return
		}
		udpconn.Write([]byte(str))
	}
}
