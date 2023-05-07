// server.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type UserConnected struct {
	Username string
	Conn     *websocket.Conn
}

var upgrader = websocket.Upgrader{} // use default options
var users [10]UserConnected
var clients map[*websocket.Conn]bool

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	clients[conn] = true

	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	// The event loop
	for {
		mt, message, err := conn.ReadMessage()

		if err != nil || mt == websocket.CloseMessage {
			break // Exit the loop if the client tries to close the connection or the connection with the interrupted client
		}

		// Send messages to all clients
		go writeMessage(message)

		go messageHandler(message)
	}

	clients[conn] = false
	defer conn.Close()
}

func messageHandler(message []byte) {
	fmt.Println(string(message))
}
func writeMessage(message []byte) {
	for conn := range clients {
		conn.WriteMessage(websocket.TextMessage, message)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Index Page")
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error env")
	}

	clients = make(map[*websocket.Conn]bool)
	http.HandleFunc("/socket", socketHandler)
	http.HandleFunc("/", home)
	fmt.Println("Server is running..." + os.Getenv("HOST") + ":" + os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(os.Getenv("HOST")+":"+os.Getenv("PORT"), nil))

}
