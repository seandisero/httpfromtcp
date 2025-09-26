package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/seandisero/httpfromtcp/internal/headers"
)

const (
	crlf       = "\r\n"
	bufferSize = 8
)

type RequestState int

const (
	requestStateInitialized RequestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine    RequestLine
	Headers        headers.Headers
	Body           []byte
	state          RequestState
	bodyLengthRead int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network conumBytesParsedection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		req, idx, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if idx == 0 {
			// need more data
			return 0, nil
		}
		r.RequestLine = *req
		r.state = requestStateParsingHeaders
		return idx, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		contentLenStr, ok := r.Headers.Get("Content-Length")
		if !ok {
			// assume that if no content-length header is present, there is no body
			r.state = requestStateDone
			return len(data), nil
		}
		contentLen, err := strconv.Atoi(contentLenStr)
		if err != nil {
			return 0, fmt.Errorf("malformed Content-Length: %s", err)
		}
		r.Body = append(r.Body, data...)
		r.bodyLengthRead += len(data)
		if r.bodyLengthRead > contentLen {
			return 0, fmt.Errorf("Content-Length too large")
		}
		if r.bodyLengthRead == contentLen {
			r.state = requestStateDone
		}
		return len(data), nil
	case requestStateDone:
		return 0, fmt.Errorf("parsing is done")
	default:
		return 0, fmt.Errorf("unknown state")
	}
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}

		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
		state:   requestStateInitialized,
	}

	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != requestStateDone {
					return nil, fmt.Errorf("incompete request, in state %d, read n bytes on EOF: %d", req.state, numBytesRead)
				}
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}
	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	reqLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return reqLine, idx + 2, nil
}

func requestLineFromString(line string) (*RequestLine, error) {
	split := strings.Split(line, " ")
	if len(split) < 3 {
		return nil, fmt.Errorf("wrong number of splits in request line")
	}

	if strings.ToUpper(split[0]) != split[0] {
		return nil, fmt.Errorf("method must be all uppercase")
	}

	versionSplit := strings.Split(split[2], "/")
	if len(versionSplit) != 2 {
		return nil, fmt.Errorf("malformed request line")
	}

	if versionSplit[1] != "1.1" {
		return nil, fmt.Errorf("wrong http version")
	}

	reqLine := RequestLine{
		HttpVersion:   versionSplit[1],
		RequestTarget: split[1],
		Method:        split[0],
	}

	return &reqLine, nil
}
