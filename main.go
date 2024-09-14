package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

var caster = "caster.centipede.fr:2101"

func main() {
	// Get the parameters from the command line
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <listening port> <caster address>", os.Args[0])
		return
	}
	address := "0.0.0.0:" + os.Args[1]
	caster = os.Args[2]

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()
	log.Printf("NTRIP Caster listening on %s", address)
	log.Printf("Forwarding requests to %s", caster)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	request, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading request: %v", err)
		return
	}

	if strings.HasPrefix(request, "GET") {
		handleGetRequest(conn, request)
	} else {
		log.Printf("Invalid request: %s", request)
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
	}
}

func handleGetRequest(conn net.Conn, request string) {
	log.Printf("Received GET request: %s", request)
	mountPointName := extractMountPointName(request)

	casterSocket, err := net.Dial("tcp", caster)
	if err != nil {
		log.Printf("Error connecting to caster: %v", err)
		return
	}
	defer casterSocket.Close()

	getRequest := fmt.Sprintf("GET /%s HTTP/1.0\r\nHost: %s\r\nUser-Agent: NTRIP proxy/0.0.0\r\nAuthorization: Basic Y2VudGlwZWRlOmNlbnRpcGVkZQ==\r\n\r\n", mountPointName, caster)
	_, err = casterSocket.Write([]byte(getRequest))
	if err != nil {
		log.Printf("Error sending GET request: %v", err)
		return
	}

	_, err = io.Copy(conn, casterSocket)
	if err != nil {
		log.Printf("Error transferring data: %v", err)
		return
	}
}

func extractMountPointName(request string) string {
	parts := strings.Split(request, " ")
	if len(parts) < 2 {
		return ""
	}
	return strings.Trim(parts[1], "/")
}
