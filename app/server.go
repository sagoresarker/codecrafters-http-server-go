package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type httpStatusCodes struct {
	OK        []byte
	NOT_FOUND []byte
}

var HttpStatus httpStatusCodes = httpStatusCodes{
	OK:        []byte("HTTP/1.1 200 OK\r\n"),
	NOT_FOUND: []byte("HTTP/1.1 404 Not Found\r\n\r\n"),
}

type httpMethods struct {
	GET  string
	POST string
}

var HttpMethod = httpMethods{
	GET:  "GET",
	POST: "POST",
}

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
}

func extractRequest(conn net.Conn) (*Request, error) {
	msg := make([]byte, 1024)
	reqLen, err := conn.Read(msg)
	if err != nil {
		return nil, err
	}

	rawReq := string(msg[:reqLen])
	lines := strings.Split(rawReq, "\r\n")
	requestLine := strings.Split(lines[0], " ")
	method := requestLine[0]
	path := requestLine[1]

	headers := make(map[string]string)
	for i := 1; i < len(lines)-2; i++ {
		headerParts := strings.Split(lines[i], ": ")

		if len(headerParts) == 2 {
			headers[headerParts[0]] = headerParts[1]
		}
	}
	return &Request{
		Method:  method,
		Path:    path,
		Headers: headers,
	}, nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	req, err := extractRequest(conn)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		return
	}

	if req.Path == "/user-agent" {
		userAgent, ok := req.Headers["User-Agent"]
		if !ok {
			userAgent = ""
		}
		contentLengthHeader := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(userAgent))
		response := append(HttpStatus.OK, []byte("Content-Type: text/plain\r\n")...)
		response = append(response, []byte(contentLengthHeader)...)
		response = append(response, []byte(userAgent)...)

		_, err = conn.Write(response)
		if err != nil {
			fmt.Println("Error writing response: ", err.Error())
			return
		}
	} else if strings.HasPrefix(req.Path, "/echo/") {
		echoContent := strings.TrimPrefix(req.Path, "/echo/")
		contentLengthHeader := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(echoContent))
		response := append(HttpStatus.OK, []byte("Content-Type: text/plain\r\n")...)
		response = append(response, []byte(contentLengthHeader)...)
		response = append(response, []byte(echoContent)...)

		_, err = conn.Write(response)
		if err != nil {
			fmt.Println("Error writing response: ", err.Error())
			return
		}
	} else {
		switch req.Path {
		case "/":
			_, err = conn.Write(append(HttpStatus.OK, []byte("\r\n")...))
			if err != nil {
				fmt.Println("Error writing response: ", err.Error())
				return
			}
		default:
			_, err = conn.Write(HttpStatus.NOT_FOUND)
			if err != nil {
				fmt.Println("Error writing response: ", err.Error())
				return
			}
		}
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)

	}

}
