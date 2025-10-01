package server

import (
	"io"

	"github.com/seandisero/httpfromtcp/internal/request"
	"github.com/seandisero/httpfromtcp/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)

func (he *HandlerError) Write(w io.ReadWriteCloser) {
	// body := []byte(he.Message)
	// h := response.GetDefaultHeaders(len(body))
	// err := response.WriteStatusLine(w, he.StatusCode)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if err = response.WriteHeaders(w, h); err != nil {
	// 	log.Fatal(err)
	// }
	// _, err = w.Write(body)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
