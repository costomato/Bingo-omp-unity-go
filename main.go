package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"util/util"

	"github.com/gorilla/websocket"
)

type RoomResponse struct {
	Channel   string `json:"channel"`
	Res       string `json:"res" default:""`
	RoomCode  string `json:"roomCode" default:""`
	Dimension int    `json:"dimension" default:"0"`
	IsCreator bool   `json:"isCreator" default:"false"`
	Move      int    `json:"move" default:"0"`
}

type Player struct {
	Name   string `default:""`
	Socket *websocket.Conn
}
type Room struct {
	Creator   Player
	Joiner    Player
	Dimension int
}

var rooms = make(map[string]Room)

func createRoom(roomCode string, creator Player, dimension int) {
	rooms[roomCode] = Room{Creator: creator, Dimension: dimension}
}
func joinRoom(roomCode string, joiner Player) {
	if entry, ok := rooms[roomCode]; ok {
		entry.Joiner = joiner
		rooms[roomCode] = entry
	}
}
func getRoom(roomCode string) Room {
	return rooms[roomCode]
}

var upgrader = websocket.Upgrader{}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
	// http.ServeFile(w, r, "./public")
}
func reader(conn *websocket.Conn) {
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var data RoomResponse
		json.Unmarshal([]byte(p), &data)

		switch data.Channel {
		case "create-room":
			fmt.Println("Creating room...")
			roomCode := util.GenerateRoomCode(5)

			fmt.Println("Room code: ", roomCode)
			fmt.Println("Creator: ", data.Res)
			createRoom(roomCode, Player{Name: data.Res, Socket: conn}, data.Dimension)

			msg, _ := json.Marshal(&RoomResponse{Channel: "create-room", Res: roomCode, RoomCode: roomCode})
			conn.WriteMessage(messageType, msg)

		case "join-room":
			fmt.Println("Creating room...")
			room := getRoom(data.RoomCode)
			joinRoom(data.RoomCode, Player{Name: data.Res, Socket: conn})

			msgToJoiner, _ := json.Marshal(&RoomResponse{Channel: "game-ready", Res: room.Creator.Name, Dimension: room.Dimension, IsCreator: false})
			conn.WriteMessage(messageType, msgToJoiner)

			msgToCreator, _ := json.Marshal(&RoomResponse{Channel: "game-ready", Res: room.Joiner.Name, IsCreator: true})
			room.Creator.Socket.WriteMessage(messageType, msgToCreator)

			fmt.Println("Game is ready")

		case "game-on":
			room := getRoom(data.RoomCode)
			msg, _ := json.Marshal(&RoomResponse{Channel: "game-on", Move: data.Move})
			if data.IsCreator {
				room.Creator.Socket.WriteMessage(messageType, msg)
			} else {
				room.Joiner.Socket.WriteMessage(messageType, msg)
			}

		case "win-claim":
			room := getRoom(data.RoomCode)
			msg, _ := json.Marshal(&RoomResponse{Channel: "win-claim"})
			if data.IsCreator {
				room.Creator.Socket.WriteMessage(messageType, msg)
			} else {
				room.Joiner.Socket.WriteMessage(messageType, msg)
			}

		case "retry":
			room := getRoom(data.RoomCode)
			msg, _ := json.Marshal(&RoomResponse{Channel: "retry", IsCreator: data.IsCreator})
			room.Creator.Socket.WriteMessage(messageType, msg)
			room.Joiner.Socket.WriteMessage(messageType, msg)

		default:
			fmt.Println("Channel not implemented")
		}
	}
}
func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Client Connected")
	reader(ws)
}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	setupRoutes()
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}
	fmt.Println("Server listening on port " + port + " >:)")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
