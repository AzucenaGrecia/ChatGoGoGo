// client.go
package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
)

var done chan interface{}
var interrupt chan os.Signal

func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %s\n", msg)
	}
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error env")
	}
	fmt.Println("App will connect to ..." + os.Getenv("HOST") + ":" + os.Getenv("PORT"))

	done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	socketUrl := "ws://" + os.Getenv("HOST") + ":" + os.Getenv("PORT") + "/socket"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()
	go receiveHandler(conn)

	// Our main loop for the client
	// We send our relevant packets here

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Text to send: ")

		input, _ := reader.ReadString('\n')
		err_socket := conn.WriteMessage(websocket.TextMessage, []byte(input))
		if err_socket != nil {
			log.Println("Error during writing to websocket:", err_socket)
		}

	}
}
