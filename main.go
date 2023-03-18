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
	Channel    string `json:"channel"`
	Res        string `json:"res" default:""`
	RoomCode   string `json:"roomCode" default:""`
	Dimension  int    `json:"dimension" default:"0"`
	IsCreator  bool   `json:"isCreator" default:"false"`
	Move       int    `json:"move" default:"0"`
	AppVersion string `json:"appVersion" default:""`
}

type Player struct {
	Name   string `default:""`
	Socket *websocket.Conn
}
type Room struct {
	Creator    Player
	Joiner     Player
	Dimension  int
	AppVersion string
}

var rooms = make(map[string]Room)

func createRoom(roomCode string, creator Player, dimension int, appVersion string) {
	rooms[roomCode] = Room{Creator: creator, Dimension: dimension, AppVersion: appVersion}
}
func joinRoom(roomCode string, joiner Player) {
	if entry, ok := rooms[roomCode]; ok {
		entry.Joiner = joiner
		rooms[roomCode] = entry
	}
}
func getRoom(roomCode string) (Room, bool) {
	if _, ok := rooms[roomCode]; ok {
		return rooms[roomCode], false
	}
	return Room{}, true
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

			createRoom(roomCode, Player{Name: data.Res, Socket: conn}, data.Dimension, data.AppVersion)

			msg, _ := json.Marshal(&RoomResponse{Channel: "create-room", Res: roomCode, RoomCode: roomCode})
			conn.WriteMessage(messageType, msg)

		case "join-room":
			fmt.Println("Creating room...")
			room, error := getRoom(data.RoomCode)
			if error {
				msgToJoiner, _ := json.Marshal(&RoomResponse{Channel: "error", Res: "The room code you entered is invalid"})
				conn.WriteMessage(messageType, msgToJoiner)
			} else {
				if room.AppVersion == data.AppVersion {
					if room.Joiner.Name == "" {
						joinRoom(data.RoomCode, Player{Name: data.Res, Socket: conn})

						msgToJoiner, _ := json.Marshal(&RoomResponse{Channel: "game-ready", Res: room.Creator.Name, Dimension: room.Dimension, IsCreator: false})
						conn.WriteMessage(messageType, msgToJoiner)

						msgToCreator, _ := json.Marshal(&RoomResponse{Channel: "game-ready", Res: data.Res, IsCreator: true})
						room.Creator.Socket.WriteMessage(messageType, msgToCreator)
					} else {
						msgToJoiner, _ := json.Marshal(&RoomResponse{Channel: "error", Res: "Room is already full"})
						conn.WriteMessage(messageType, msgToJoiner)
					}
				} else {
					msgToJoiner, _ := json.Marshal(&RoomResponse{Channel: "error", Res: "Room creator has a different version of Bingo. Please make sure both have the latest version."})
					conn.WriteMessage(messageType, msgToJoiner)
				}
			}

			fmt.Println("Game is ready")

		case "game-on":
			room, _ := getRoom(data.RoomCode)
			msg, _ := json.Marshal(&RoomResponse{Channel: "game-on", Move: data.Move})
			if data.IsCreator {
				room.Creator.Socket.WriteMessage(messageType, msg)
			} else {
				room.Joiner.Socket.WriteMessage(messageType, msg)
			}

		case "win-claim":
			room, _ := getRoom(data.RoomCode)
			msg, _ := json.Marshal(&RoomResponse{Channel: "win-claim"})
			if data.IsCreator {
				room.Creator.Socket.WriteMessage(messageType, msg)
			} else {
				room.Joiner.Socket.WriteMessage(messageType, msg)
			}

		case "retry":
			room, _ := getRoom(data.RoomCode)
			msg, _ := json.Marshal(&RoomResponse{Channel: "retry"})
			if data.IsCreator {
				room.Creator.Socket.WriteMessage(messageType, msg)
			} else {
				room.Joiner.Socket.WriteMessage(messageType, msg)
			}

		case "exit-room":
			room, _ := getRoom(data.RoomCode)
			delete(rooms, data.RoomCode)
			msg, _ := json.Marshal(&RoomResponse{Channel: "exit-room"})
			if data.IsCreator {
				room.Creator.Socket.WriteMessage(messageType, msg)
			} else {
				room.Joiner.Socket.WriteMessage(messageType, msg)
			}

		default:
			fmt.Println("Channel not implemented")
		}
	}
}
func wsEndpoint(w http.ResponseWriter, r *http.Request) {
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
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	setupRoutes()
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			next.ServeHTTP(w, r)
		})
	}
	log.Fatal(http.ListenAndServe(":"+port, corsHandler(http.DefaultServeMux)))
	fmt.Println("Server listening on port " + port + " >:)")
}
