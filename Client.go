// client.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/exec"
	"os/signal"
)

var done chan interface{}
var interrupt chan os.Signal
var clien_user_name string
var mode string

type SocketMessage struct {
	Username    string
	MessageType int // 1. Login // 2 Chat // 3 list users in chat
	Text        string
}

type SocketResponse struct {
	MessageType int // 1. Login // 2 Chat // 3 list users in chat
	Text        string
}

func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}

		var response SocketResponse
		json.Unmarshal(msg, &response)

		switch response.MessageType {
		case 2:
			if mode == "CHAT" {
				fmt.Println(response.Text)
			}
		case 3:
			fmt.Println(response.Text)
			printMenu()
		}

	}
}

func welcome_and_login() string {
	fmt.Println("Hola Bienvenido al chat")
	fmt.Println("Ingresa tu nombre de usuario:")
	reader := bufio.NewReader(os.Stdin)
	userName, _ := reader.ReadString('\n')

	return userName
}

func login(conn *websocket.Conn) {
	message := SocketMessage{
		Username:    clien_user_name,
		MessageType: 1,
		Text:        "",
	}
	bytes_message, _ := json.Marshal(message)
	err_socket := conn.WriteMessage(websocket.TextMessage, bytes_message)
	if err_socket != nil {
		log.Println("Error during writing to websocket:", err_socket)
	}
	fmt.Println("Logged in the chat ")
}

func sendMessage(conn *websocket.Conn, text string) {
	message := SocketMessage{
		Username:    clien_user_name,
		MessageType: 2,
		Text:        text,
	}

	bytes_message, _ := json.Marshal(message)
	err_socket := conn.WriteMessage(websocket.TextMessage, bytes_message)
	if err_socket != nil {
		log.Println("Error during writing to websocket:", err_socket)
	}
}

func listUsers(conn *websocket.Conn) {
	message := SocketMessage{
		Username:    clien_user_name,
		MessageType: 3,
		Text:        "",
	}
	bytes_message, _ := json.Marshal(message)
	err_socket := conn.WriteMessage(websocket.TextMessage, bytes_message)
	if err_socket != nil {
		log.Println("Error during writing to websocket:", err_socket)
	}
}

func makeCommand(command string) {
	cmd := exec.Command(command) //Linux example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func printMenu() {
	fmt.Println("Ejecuta un comando:")
	fmt.Println("1. Listar usuarios")
	fmt.Println("2. Chat en vivo")
}
func main() {
	mode = "MENU"
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error env")
	}
	reader := bufio.NewReader(os.Stdin)

	username := welcome_and_login()
	clien_user_name = username

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

	//LOGIN TO CHAT
	login(conn)

	for {
		if mode == "MENU" {
			printMenu()
			input, _ := reader.ReadString('\n')
			command := input[0:1]

			switch command {
			case "1":
				fmt.Println("Listar usuarios")
				listUsers(conn)
			case "2":
				mode = "CHAT"
			default:
				fmt.Println("No command")
			}
		}
		if mode == "CHAT" {
			fmt.Println("Text to send:  (Q to quit)")
			input, _ := reader.ReadString('\n')
			command := input[0:2]
			if command == "q\n" {
				mode = "MENU"
				makeCommand("clear")
			} else {
				sendMessage(conn, input)
			}
		}

	}
}
