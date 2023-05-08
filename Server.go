// server.go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

type UserConnected struct {
	Username string
	Conn     *websocket.Conn
	Online   bool
	exist    bool
}
type SocketMessageServer struct {
	Username    string
	MessageType int // 1. Login // 2 Chat
	Text        string
}
type SocketResponseServer struct {
	MessageType int // 1. Login // 2 Chat // 3 list users in chat
	Text        string
}

var upgrader = websocket.Upgrader{} // use default options
var users [100]UserConnected
var clients map[*websocket.Conn]bool

func findUser(username string) UserConnected {
	for _, user := range users {
		if user.Username == username {
			return user
		}
	}
	return UserConnected{}
}

func list_all_users() string {
	list := "Usuarios conectados: \n"
	for _, user := range users {
		if user.Online {
			list += user.Username + "\n"
		}
	}

	return list
}

func logUser(conn *websocket.Conn, username string) bool {

	for i := 0; i < 100; i++ {
		if !users[i].exist {
			users[i] = UserConnected{
				Username: username,
				Conn:     conn,
				Online:   true,
				exist:    true,
			}

			fmt.Println("Usuario registrado")
			return true
		}
	}

	return false
}

func validateSocketMessage(conn *websocket.Conn, message []byte) {
	var recived_message SocketMessageServer
	err_message_recived := json.Unmarshal(message, &recived_message)
	if err_message_recived != nil {
		log.Fatal("ERROR READING MESSAGE")
	}
	switch recived_message.MessageType {
	case 1:
		logUser(conn, recived_message.Username)

		fmt.Println(users)
		fmt.Println("Userr logger " + recived_message.Username)
	case 2: // Send messages to all clients
		response := SocketResponseServer{
			MessageType: 2,
			Text:        recived_message.Username + " -> " + recived_message.Text,
		}
		go writeMessage(response)
		go messageHandler(response)
	case 3:
		response := SocketResponseServer{
			MessageType: 3,
			Text:        list_all_users(),
		}
		writeMessageToUser(response, recived_message.Username)

	default:
		fmt.Println("No action recived")
	}
}
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

		//UNMARSHALL MESSAGE
		validateSocketMessage(conn, message)

	}

	clients[conn] = false
	conn.Close()
}

func messageHandler(response SocketResponseServer) {
	fmt.Println(response.Text)
}
func writeMessage(response SocketResponseServer) {
	bytes_message, _ := json.Marshal(response)

	for _, user := range users {
		if user.Online {
			user.Conn.WriteMessage(websocket.TextMessage, bytes_message)
		}
	}
}

func writeMessageToUser(response SocketResponseServer, username string) {
	user_found := findUser(username)
	bytes_message, _ := json.Marshal(response)
	user_found.Conn.WriteMessage(websocket.TextMessage, bytes_message)

}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Index Page")
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error env")
	}
	for i := 0; i < 100; i++ {
		users[i] = UserConnected{}
	}

	clients = make(map[*websocket.Conn]bool)
	http.HandleFunc("/socket", socketHandler)
	http.HandleFunc("/", home)
	fmt.Println("Server is running..." + os.Getenv("HOST") + ":" + os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(os.Getenv("HOST")+":"+os.Getenv("PORT"), nil))

}
