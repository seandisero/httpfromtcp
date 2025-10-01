package response

import (
	"fmt"
	"io"

	"github.com/seandisero/httpfromtcp/internal/headers"
)

type Writer struct {
	writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		writer: writer,
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	hdrs := headers.NewHeaders()
	hdrs.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	hdrs.Set("Connection", "close")
	hdrs.Set("Content-Type", "text/plain")

	return hdrs
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var statusLine = []byte{}
	switch statusCode {
	case StatusOK:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		return fmt.Errorf("undefined status code behaviour")
	}

	_, err := w.writer.Write(statusLine)
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	data := []byte{}
	for key, value := range headers {
		data = fmt.Append(data, key, ": ", value, "\r\n")
	}
	data = fmt.Append(data, "\r\n")
	_, err := w.writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	return w.WriteBody(p)
}
func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.WriteBody([]byte("\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	data := []byte{}
	for key, value := range h {
		data = fmt.Append(data, key, ": ", value, "\r\n")
	}
	data = fmt.Append(data, "\r\n")
	_, err := w.writer.Write(data)
	if err != nil {
		return err
	}

	return nil
}
