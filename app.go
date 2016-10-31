package main

import (
	"github.com/Coffee-Lovers/WSServer/messages"
	"github.com/Coffee-Lovers/WSServer/wsserver"
	"log"
	"encoding/json"
	"github.com/Coffee-Lovers/WSServer/config"
)

type App struct {
	connections chan *wsserver.ConnectionDescription
	messages    chan messages.Progress
}

func (app *App) run() {
	app.startWebSocketServer()
	app.startHubProgressListener()

	app.loop()
}

func (app *App) startHubProgressListener() {
	app.messages = make(chan messages.Progress)
	hub := messages.RabbitMQ{
		Username: config.Rabbit["user"],
		Password: config.Rabbit["pass"],
		Host: config.Rabbit["host"],
		Port: config.Rabbit["port"],
	}
	go hub.Listen(app.messages)
}

func (app *App) startWebSocketServer() {
	app.connections = make(chan *wsserver.ConnectionDescription)
	wss := wsserver.WSServer{
		Port: config.WSS["port"],
		Path: config.WSS["path"],
	}
	go wss.RunWSServer(app.connections)
}

func (app *App) loop() {
	log.Println("Waiting for something to do...")
	router := NewRouter()
	for {
		select {
		case connection := <-app.connections:
			log.Printf("New websocket connection arrived - %+v \n", connection)
			router.Join(connection)

		case message := <-app.messages:
			log.Printf("New message from hub arrived - %+v \n", message)
			router.MessageArrived(message)
		}
	}
}

func main() {
	app := &App{}
	app.run()
}

type Router struct {
	connections []*wsserver.ConnectionDescription
	messages []messages.Progress
}

func (r *Router) Join(cd *wsserver.ConnectionDescription)  {
	r.connections = append(r.connections, cd)
	// check if there are messages waiting for this connection
	for _, m := range r.messages {
		if id, _ := m.GetRelatedTaskId(); id == cd.SubscribeTo {
			cd.SendChannel <- toByte(m)
		}
	}
}

func (r *Router) MessageArrived(m messages.Progress)  {
	r.messages = append(r.messages, m)
	// send to all interested connections
	for _, c := range r.connections {
		if id, _ :=  m.GetRelatedTaskId(); c.SubscribeTo == id {
			c.SendChannel <- toByte(m)
		}
	}
}

func toByte(m messages.Progress) []byte {
	b, _ := json.Marshal(m)
	return b
}

func NewRouter() Router {
	return Router{
		make([]*wsserver.ConnectionDescription, 0),
		make([]messages.Progress, 0),
	}
}


