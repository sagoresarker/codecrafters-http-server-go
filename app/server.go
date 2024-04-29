package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	path2 "path"
	"strings"
)

type Request struct {
	Method    string
	Path      string
	Protocol  string
	Host      string
	UserAgent string
	Body      string
}

const (
	CRLF             = "\r\n"
	OK               = "HTTP/1.1 200 OK"
	NotFound         = "HTTP/1.1 404 Not Found"
	BadRequest       = "HTTP/1.1 400 Bad Request"
	InternalError    = "HTTP/1.1 500 Internal Server Error"
	ContentTypeText  = "Content-Type: text/plain"
	ContentTypeOctet = "Content-Type: application/octet-stream"
	ContentLength    = "Content-Length:"
	Created          = "HTTP/1.1 201 Created"
)

var (
	directory string
)

func newRequest(method string, path string, protocol string, host string, userAgent string, body string) *Request {
	return &Request{
		Method:    method,
		Path:      path,
		Protocol:  protocol,
		Host:      host,
		UserAgent: userAgent,
		Body:      body,
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		log.Printf("Error reading: %v", err)
		return
	}

	buf = bytes.Trim(buf, "\x00")
	reqString := string(buf[:n])
	reqStringSlice := strings.Split(reqString, CRLF)

	if len(reqStringSlice) < 3 {
		sendResponse(conn, BadRequest, "")
		return
	}

	startLineSlice := strings.Split(reqStringSlice[0], " ")
	if len(startLineSlice) < 3 {
		log.Println("Invalid start line in request")
		sendResponse(conn, BadRequest, "")
		return
	}

	headers := parseHeaders(reqStringSlice[1:])
	host := headers["host"]
	userAgent := headers["user-agent"]
	body := reqStringSlice[len(reqStringSlice)-1]

	request := newRequest(startLineSlice[0], startLineSlice[1], startLineSlice[2], host, userAgent, body)

	fmt.Printf("New Request: %s %s %s\n", request.Method, request.Path, request.Protocol)

	switch {
	case request.Path == "/":
		sendResponse(conn, OK, "")
	case strings.HasPrefix(request.Path, "/echo"):
		handleEcho(conn, request)
	case request.Path == "/user-agent":
		handleUserAgent(conn, request)
	case strings.HasPrefix(request.Path, "/files"):
		if request.Method == "POST" {
			handleUploadFile(conn, request)
			break
		}
		handleDownLoadFile(conn, request)
	default:
		sendResponse(conn, NotFound, "")
	}
}

func parseHeaders(lines []string) map[string]string {
	headers := make(map[string]string)
	for _, line := range lines {
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			headers[strings.ToLower(parts[0])] = parts[1]
		}
	}
	return headers
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	flag.StringVar(&directory, "directory", "files", "path to the file directory")
	flag.Parse()
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Failed to accept connection")
		}
		go handleConnection(conn)
	}
}

func handleEcho(conn net.Conn, request *Request) {
	body, found := strings.CutPrefix(request.Path, "/echo/")
	if !found {
		fmt.Println("Failed to parse request")
		handleServerError(conn)
		return
	}
	sendResponse(conn, OK, body)
}

func handleDownLoadFile(conn net.Conn, request *Request) {
	fileName, found := strings.CutPrefix(request.Path, "/files/")
	if !found {
		fmt.Println("Failed to parse request")
		handleServerError(conn)
		return
	}
	path := path2.Join(directory, fileName)
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Failed to open file \t" + path)
		sendResponse(conn, NotFound, "")
		return
	}
	downloadFile(conn, OK, file)
}

func handleUploadFile(conn net.Conn, request *Request) {
	fileName, found := strings.CutPrefix(request.Path, "/files/")
	if !found {
		fmt.Println("Failed to parse request")
		handleServerError(conn)
		return
	}
	path := path2.Join(directory, fileName)
	err := os.WriteFile(path, []byte(request.Body), 0644)
	if err != nil {
		handleServerError(conn)
		return
	}
	saveFile(conn)
}

func handleUserAgent(conn net.Conn, request *Request) {
	sendResponse(conn, OK, request.UserAgent)
}

func handleServerError(conn net.Conn) {
	sendResponse(conn, InternalError+CRLF+CRLF, "")
}

func sendResponse(conn net.Conn, status string, body string) {
	response := fmt.Sprintf("%s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", status, len(body), body)
	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		handleServerError(conn)
	}
}

func downloadFile(conn net.Conn, status string, body []byte) {
	response := fmt.Sprintf("%s\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", status, len(body), body)
	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error downloading file: %v", err)
		handleServerError(conn)
	}
}

func saveFile(conn net.Conn) {
	_, err := conn.Write([]byte(Created + CRLF + CRLF))
	if err != nil {
		log.Printf("Error downloading file: %v", err)
		handleServerError(conn)
	}
}
