package wsserver

import (
	"net/http"
	"time"
	"github.com/gorilla/websocket"
	"log"
)

type ConnectionDescription struct {
	SubscribeTo string
	SendChannel chan []byte
	connection *websocket.Conn
}

type WSServer struct {
	Path string
	Port string
}

func (wss WSServer) RunWSServer(connectionsChan chan *ConnectionDescription) {
	s := &http.Server{
		Addr:           ":" + wss.Port,
		ReadTimeout:    1 * time.Second,
		WriteTimeout:   1 * time.Second,
	}

	http.HandleFunc(wss.Path, getWSConnectionHandler(connectionsChan))

	go s.ListenAndServe()
}

func getWSConnectionHandler(c chan *ConnectionDescription) func (w http.ResponseWriter, r *http.Request) {

	return func (w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		con, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		if conDesc, err := processSubscriber(con); err == nil {
			c <- conDesc
		} else {
			log.Println("Failed to establish connection:", err)
			con.Close()
		}
	}
}

func processSubscriber(con *websocket.Conn) (*ConnectionDescription, error) {
	cd := &ConnectionDescription{}
	if err := con.ReadJSON(cd); err == nil {
		cd.connection = con
		cd.SendChannel = make(chan []byte)

		// listen for something to write
		go func(cd *ConnectionDescription) {
			for m := range cd.SendChannel {
				cd.connection.WriteMessage(websocket.TextMessage, m)
			}
		}(cd)

		return cd, nil
	} else {
		return &ConnectionDescription{}, err
	}
}
