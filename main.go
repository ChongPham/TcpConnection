package main

import (
	"io"
	"log"
	"net"
	"time"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Println(conn.RemoteAddr())
	//read data from client
	var buffer []byte = make([]byte, 1000)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Println("Client closed connection:", conn.RemoteAddr())
				return
			}
			log.Println("Read error:", err)
			return
		}

		req := string(buffer[:n])
		log.Println("Request from", conn.RemoteAddr(), ":\n", req)
		//process sometihing
		time.Sleep(time.Second * 2)

		//rely to client
		response := "HTTP/1.1 200 OK\r\nContent-Length: 12\r\n\r\nhello world!"
		m, err := conn.Write([]byte(response))
		if err != nil {
			log.Println("Write error:", err)
			return
		}
		log.Println("Request tiep theo ne", m)
	}
}

func main() {
	// tao tcp socket lang nghe o port 3000
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	for {
		// thiet lap dedicated connection channel
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		log.Println(conn)

		// create goroutine to handles
		go handleConnection(conn)
	}
}
