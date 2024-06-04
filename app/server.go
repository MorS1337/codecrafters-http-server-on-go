package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")
	
	ln, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	
	defer ln.Close()
	
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go HandleConnection(conn)
	}
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		return
	}
	var response string

	if strings.HasPrefix(req.URL.Path, "/echo/") {
		echoStr := strings.TrimPrefix(req.URL.Path, "/echo/")
		contentEncStr := CheckEncoding(req)
		fmt.Print(len(req.Header))
		if contentEncStr != "invalid-encoding" && contentEncStr != "" {
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\nContent-Encoding: %s\r\n\r\n%s", len(echoStr), contentEncStr, echoStr)
		} else {
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echoStr), echoStr)
		}
	} else if strings.HasPrefix(req.URL.Path, "/user-agent") {
		uaStr := req.Header["User-Agent"][0]
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(uaStr), uaStr)
	} else if strings.HasPrefix(req.URL.Path, "/files/") {
		dir := os.Args[2]
		fileName := strings.TrimPrefix(req.URL.Path, "/files/")
		if req.Method == "POST" {
			content, _ := ConvertBody(req.Body)
			err := os.WriteFile(dir + fileName, []byte(content), 0644)
			if err != nil {
				fmt.Println("Error writing file: ", err.Error())
			}
			response = "HTTP/1.1 201 Created\r\n\r\n"
		} else {
		data, err := os.ReadFile(dir + fileName)
		if err != nil {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
		} else {
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(data), data)
		}
	}
	} else if req.URL.Path == "/" {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
	}
}

func ConvertBody(body io.ReadCloser) (string, error) {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, body)
	return buf.String(), err
}

func CheckEncoding(req *http.Request) string {
	if len(req.Header) > 1 {
		return req.Header["Accept-Encoding"][0]
	}
	return ""
}