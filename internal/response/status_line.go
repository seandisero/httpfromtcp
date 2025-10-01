package response

import "io"

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case StatusOK:
		w.Write([]byte("HTTP/1.1 200 OK\r\n"))
		return nil
	case StatusBadRequest:
		w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		return nil
	case StatusInternalServerError:
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		return nil
	default:
		return nil
	}
}
