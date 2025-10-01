package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/seandisero/httpfromtcp/internal/headers"
	"github.com/seandisero/httpfromtcp/internal/request"
	"github.com/seandisero/httpfromtcp/internal/response"
	"github.com/seandisero/httpfromtcp/internal/server"
)

const port = 42069

const badReq = `
<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

const svrErr = `
<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
	`

const isOk = `
<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
	`

func toStr(hash []byte) string {
	out := ""
	for _, b := range hash {
		out += fmt.Sprintf("%02x", b)
	}
	return out
}

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		h.Replace("Content-Type", "text/html")
		requestTarget := req.RequestLine.RequestTarget
		if requestTarget == "/yourproblem" {
			w.WriteStatusLine(response.StatusBadRequest)

			body := []byte(badReq)
			h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
			w.WriteHeaders(h)
			w.WriteBody(body)
		} else if requestTarget == "/myproblem" {
			w.WriteStatusLine(response.StatusInternalServerError)

			body := []byte(svrErr)
			h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
			w.WriteHeaders(h)
			w.WriteBody(body)
		} else if strings.HasPrefix(requestTarget, "/httpbin/") {
			chunkNumber := requestTarget[len("/httpbin/"):]

			h.Remove("Content-Length")
			h.Set("Transfer-Encoding", "chunked")
			h.Set("Trailer", "x-content-sha256")
			h.Set("Trailer", "x-content-length")
			url := fmt.Sprintf("https://httpbin.org/%s", chunkNumber)
			fmt.Println(url)
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println(err)
				return
			}
			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(h)
			data := []byte{}
			for {
				var p = make([]byte, 32)
				n, err := resp.Body.Read(p)
				if err != nil {
					fmt.Println(err)
					break
				}
				fmt.Printf("data size: %d\n", n)
				w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
				w.WriteBody(p[:n])
				w.WriteBody([]byte("\r\n"))
				data = append(data, p[:n]...)
			}
			w.WriteBody([]byte("0\r\n"))
			trailers := headers.NewHeaders()
			fmt.Printf("full body length: %d\n", len(data))
			hash := sha256.Sum256(data)
			sha256Hash := fmt.Sprintf("%x", hash)
			fmt.Printf("calcualted sha256: %s\n", sha256Hash)
			trailers.Replace("X-Content-Sha256", sha256Hash)
			trailers.Replace("X-Content-Length", fmt.Sprintf("%d", len(data)))
			fmt.Printf("Trailer header: %s\n", h["trailer"])
			err = w.WriteTrailers(trailers)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
		} else if requestTarget == "/video" {
			f, err := os.ReadFile("assets/vim.mp4")
			if err != nil {
				return
			}
			h.Replace("Content-Type", "video/mp4")
			h.Replace("content-length", fmt.Sprintf("%d", len(f)))

			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(h)
			w.WriteBody(f)
		} else {
			w.WriteStatusLine(response.StatusOK)

			body := []byte(isOk)
			h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
			w.WriteHeaders(h)
			w.WriteBody(body)
		}
	})

	if err != nil {
		log.Printf("error starting server: %v", err)
	}
	defer server.Close()
	log.Println("server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("server gracefully stopped")
}
