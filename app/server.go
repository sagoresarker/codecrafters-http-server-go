package main

import (
	"fmt"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

type httpStatusCodes struct {
	OK        []byte
	NOT_FOUND []byte
}

var HttpStatus httpStatusCodes = httpStatusCodes{
	OK:        []byte("HTTP/1.1 200 OK\r\n\r\n"),
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
	Method string
	Path   string
}

func extractRequest(conn net.Conn) (*Request, error) {
	msg := make([]byte, 1024)

	reqLen, err := conn.Read(msg)

	if err != nil {

		return nil, err

	}

	rawReq := string(msg[:reqLen])

	lines := strings.Split(rawReq, "\n")

	method := strings.Split(lines[0], " ")[0]

	path := strings.Split(lines[0], " ")[1]

	return &Request{

		Method: method,

		Path: path,
	}, nil
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	req, err := extractRequest(conn)

	if err != nil {

		fmt.Println("Error reading request: ", err.Error())

		os.Exit(1)

	}

	fmt.Println(req.Path)

	switch req.Path {

	case "/":

		fmt.Println("case /")

		_, err = conn.Write(HttpStatus.OK)
		if err != nil {

			fmt.Println("Error writing response: ", err.Error())

			os.Exit(1)

		}

	default:

		fmt.Println("case 404")

		_, err = conn.Write(HttpStatus.NOT_FOUND)
		if err != nil {

			fmt.Println("Error writing response: ", err.Error())

			os.Exit(1)

		}

	}
}
