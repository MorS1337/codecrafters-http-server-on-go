package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

type Server struct {
	listener net.Listener
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	server, err := NewServer("0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to start server:", err)
		os.Exit(1)
	}

	server.Run()
}

func NewServer(address string) (*Server, error) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &Server{listener: ln}, nil
}

func (s *Server) Run() {
	defer s.listener.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go HandleConnection(conn)
	}
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}

	response := HandleRequest(req)
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

func HandleRequest(req *http.Request) string {
	switch {
	case strings.HasPrefix(req.URL.Path, "/echo/"):
		return HandleEcho(req)
	case strings.HasPrefix(req.URL.Path, "/user-agent"):
		return HandleUserAgent(req)
	case strings.HasPrefix(req.URL.Path, "/files/"):
		return HandleFiles(req)
	case req.URL.Path == "/":
		return "HTTP/1.1 200 OK\r\n\r\n"
	default:
		return "HTTP/1.1 404 Not Found\r\n\r\n"
	}
}

func HandleEcho(req *http.Request) string {
	echoStr := strings.TrimPrefix(req.URL.Path, "/echo/")
	acceptEncoding := CheckEncoding(req)

	if strings.Contains(acceptEncoding, "gzip") {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write([]byte(echoStr)); err != nil {
			fmt.Println("Error writing gzip data:", err)
			return "HTTP/1.1 500 Internal Server Error\r\n\r\n"
		}
		if err := gz.Close(); err != nil {
			fmt.Println("Error closing gzip writer:", err)
			return "HTTP/1.1 500 Internal Server Error\r\n\r\n"
		}
		gzipData := buf.Bytes()
		return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(gzipData), gzipData)
	}

	return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echoStr), echoStr)
}

func HandleUserAgent(req *http.Request) string {
	uaStr := req.Header.Get("User-Agent")
	return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(uaStr), uaStr)
}

func HandleFiles(req *http.Request) string {
	dir := os.Args[2]
		fileName := strings.TrimPrefix(req.URL.Path, "/files/")
		if req.Method == "POST" {
			content, _ := ConvertBody(req.Body)
			err := os.WriteFile(dir + fileName, []byte(content), 0644)
			if err != nil {
				fmt.Println("Error writing file: ", err.Error())
			}
		}
	return "HTTP/1.1 201 Created\r\n\r\n"
}

func CheckEncoding(req *http.Request) string {
	encodings := req.Header.Get("Accept-Encoding")
	encList := strings.Split(encodings, ", ")
	for _, enc := range encList {
		if enc == "gzip" {
			return "gzip"
		}
	}
	return ""
}

func ConvertBody(body io.ReadCloser) (string, error) {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, body)
	return buf.String(), err
}