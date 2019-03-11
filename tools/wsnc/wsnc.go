// Small utility for working with WebSocket connections on the command line.
package main

import (
	"bufio"
	"flag"
	"log"
	"os"

	"github.com/gorilla/websocket"
)

func lineReader() chan string {
	channel := make(chan string)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		defer func() {
			close(channel)
		}()

		for {
			line, err := reader.ReadString('\n')

			if err != nil {
				log.Println("Unable to read from stdin: %s", err)
				return
			}

			channel <- line
		}
	}()

	return channel
}

func messageReader(connection *websocket.Conn) chan string {
	channel := make(chan string)

	go func() {
		defer connection.Close()

		for {
			_, message, err := connection.ReadMessage()

			if err != nil {
				log.Printf("Read message failed: %s", err)
				close(channel)
				return
			}

			channel <- string(message)
		}
	}()

	return channel
}

func main() {
	log.SetFlags(0)

	server := flag.String("server", "wss://127.0.0.1:443/socket", "Specify server to connect to. For example: ws://127.0.0.1:443/socket")
	flag.Parse()

	if server == nil {
		log.Println("Missing server")
		return
	}

	log.Printf("Connecting to %s", *server)

	connection, _, err := websocket.DefaultDialer.Dial(*server, nil)

	if err != nil {
		log.Printf("Unable to connect to %s: %s", *server, err)
		return
	}

	log.Println("Connected to %s", *server)

	lineChannel := lineReader()
	messageChannel := messageReader(connection)

	for {
		select {
		case line, ok := <-lineChannel:
			if !ok {
				log.Printf("No more keyboard input")
				return
			}

			log.Printf("Writing: %s", line)
			connection.WriteMessage(websocket.TextMessage, []byte(line))
		case message, ok := <-messageChannel:
			if !ok {
				return
			}

			log.Printf("Received message: %s", message)
		}
	}
}
