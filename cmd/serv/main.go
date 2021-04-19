package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	//"os"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/stts-se/manuscriptor2000/dbapi"
	"github.com/stts-se/manuscriptor2000/filter"
	"github.com/stts-se/manuscriptor2000/protocol"
	"github.com/stts-se/manuscriptor2000/text"
)

type ClientMessage struct {
	ClientID    string `json:"client_id"`
	MessageType string `json:"message_type"`
	Payload     string `json:"payload"`
	Error       string `json:"error"`
}

type ScriptPayload struct {
	FetchOptions FetchPayload            `json:"fetch_options"`
	Metadata     protocol.ScriptMetadata `json:"metadata"`
	Script       []text.Sentence         `json:"script"`
}

type BatchPayload struct {
	FetchOptions FetchPayload           `json:"fetch_options"`
	Metadata     protocol.BatchMetadata `json:"metadata"`
	Batch        []text.Sentence        `json:"batch"`
}

type FetchPayload struct {
	Name       string `json:"name"`
	Type       string `json:"type"` // script or batch
	PageNumber int    `json:"page_number"`
	PageSize   int    `json:"page_size"`
}

var clientMutex sync.RWMutex
var clients = make(map[string]*websocket.Conn)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["client_id"]
	if clientID == "" {
		msg := "expected client ID, got empty string"
		log.Print(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("failed to upgrade HTTP request to websocket : %v", err)
	}

	clientMutex.Lock()
	clients[clientID] = ws
	clientMutex.Unlock()
	log.Printf("added client websocket for %s", clientID)

	// listen forever
	go listenToClient(ws)

}

func listenToClient(conn *websocket.Conn) {
	for {
		var msg ClientMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("websocket error : %v", err)
			return
		}

		log.Printf("%v\n", msg)

		switch msg.MessageType {
		case "list_batches":
			res, err := dbapi.ListBatches()
			if err != nil {
				log.Printf("failed to list batches: %v", err)
				return
			}
			jsn, err := json.Marshal(res)
			if err != nil {
				log.Printf("failed to marshal struct into JSON : %v", err)
				return
			}
			resp := ClientMessage{
				ClientID:    msg.ClientID,
				MessageType: "batches",
				Payload:     string(jsn),
			}

			jsnMsg, err := json.Marshal(resp)
			if err != nil {
				log.Printf("failed to marshal struct into JSON : %v", err)
				return
			}
			conn.WriteMessage(websocket.TextMessage, jsnMsg)

		case "list_scripts":
			res, err := dbapi.ListScripts()
			if err != nil {
				log.Printf("failed to list scripts: %v", err)
				return
			}
			jsn, err := json.Marshal(res)
			if err != nil {
				log.Printf("failed to marshal struct into JSON : %v", err)
				return
			}
			resp := ClientMessage{
				ClientID:    msg.ClientID,
				MessageType: "scripts",
				Payload:     string(jsn),
			}

			jsnMsg, err := json.Marshal(resp)
			if err != nil {
				log.Printf("failed to marshal struct into JSON : %v", err)
				return
			}
			conn.WriteMessage(websocket.TextMessage, jsnMsg)

		case "list_filters":
			filters := filter.AvailableFeats()
			jsnFilters, err := json.Marshal(filters)
			if err != nil {
				log.Printf("failed to marshal struct into JSON : %v", err)
				return
			}
			resp := ClientMessage{
				ClientID:    msg.ClientID,
				MessageType: "filters",
				Payload:     string(jsnFilters),
			}

			jsnMsg, err := json.Marshal(resp)
			if err != nil {
				log.Printf("failed to marshal struct into JSON : %v", err)
				return
			}
			conn.WriteMessage(websocket.TextMessage, jsnMsg)

		case "get_stats":
			stats, err := dbapi.GetStats()
			if err != nil {
				resp := ClientMessage{
					ClientID:    msg.ClientID,
					MessageType: "server_error",
					Error:       fmt.Sprintf("%v", err),
				}

				jsnMsg, err := json.Marshal(resp)
				if err != nil {
					log.Printf("failed to marshal struct into JSON : %v", err)
					return
				}

				conn.WriteMessage(websocket.TextMessage, jsnMsg)

			} else {
				jsnStats, err := json.Marshal(stats)
				if err != nil {
					log.Printf("failed to marshal struct into JSON : %v", err)
					return
				}
				resp := ClientMessage{
					ClientID:    msg.ClientID,
					MessageType: "db_stats",
					Payload:     string(jsnStats),
				}

				jsnResp, err := json.Marshal(resp)
				if err != nil {
					log.Printf("failed to marshal struct into JSON : %v", err)
					return
				}
				conn.WriteMessage(websocket.TextMessage, jsnResp)
			}
		case "block_sents":
			var sendErr error
			var ids []int
			err := json.Unmarshal([]byte(msg.Payload), &ids)
			if err != nil {
				sendErr = fmt.Errorf("failed to unmarshal ids : %v", err)
			}
			var ids64 []int64
			for _, i := range ids {
				ids64 = append(ids64, int64(i))
			}
			err = dbapi.BlockSentIDs(ids64...)
			if err != nil {
				sendErr = err
			}
			if sendErr != nil {
				resp := ClientMessage{
					ClientID:    msg.ClientID,
					MessageType: "server_error",
					Error:       fmt.Sprintf("%v", sendErr),
				}

				jsnMsg, err := json.Marshal(resp)
				if err != nil {
					log.Printf("failed to marshal struct into JSON : %v", err)
					return
				}

				conn.WriteMessage(websocket.TextMessage, jsnMsg)

			} else {
				resp := ClientMessage{
					ClientID:    msg.ClientID,
					MessageType: "blocked_sents",
					Payload:     msg.Payload,
				}

				jsnResp, err := json.Marshal(resp)
				if err != nil {
					log.Printf("failed to marshal struct into JSON : %v", err)
					return
				}
				conn.WriteMessage(websocket.TextMessage, jsnResp)
			}
		case "fetch_script":
			var sendErr error
			var opts FetchPayload
			err := json.Unmarshal([]byte(msg.Payload), &opts)
			if err != nil {
				sendErr = fmt.Errorf("failed to unmarshal opts : %v", err)
			}
			script, err := dbapi.GetScript(opts.Name, opts.PageNumber, opts.PageSize)
			if err != nil {
				sendErr = err
			}
			metadata, err := dbapi.GetScriptProperties(opts.Name)
			if err != nil {
				sendErr = err
			}

			if sendErr != nil {
				resp := ClientMessage{
					ClientID:    msg.ClientID,
					MessageType: "server_error",
					Error:       fmt.Sprintf("%v", sendErr),
				}

				jsnMsg, err := json.Marshal(resp)
				if err != nil {
					log.Printf("failed to marshal struct into JSON : %v", err)
					return
				}

				conn.WriteMessage(websocket.TextMessage, jsnMsg)

			} else {
				res := ScriptPayload{
					FetchOptions: opts,
					Metadata:     metadata,
					Script:       script,
				}
				scriptPayload, err := json.Marshal(res)
				if err != nil {
					log.Printf("failed to marshal struct into JSON : %v", err)
					return
				}
				resp := ClientMessage{
					ClientID:    msg.ClientID,
					MessageType: "fetched_script",
					Payload:     string(scriptPayload),
				}

				jsnResp, err := json.Marshal(resp)
				if err != nil {
					log.Printf("failed to marshal struct into JSON : %v", err)
					return
				}
				conn.WriteMessage(websocket.TextMessage, jsnResp)
			}
		case "fetch_batch":
			var sendErr error
			var opts FetchPayload
			err := json.Unmarshal([]byte(msg.Payload), &opts)
			if err != nil {
				sendErr = fmt.Errorf("failed to unmarshal opts : %v", err)
			}
			batch, err := dbapi.GetBatch(opts.Name, opts.PageNumber, opts.PageSize)
			if err != nil {
				sendErr = err
			}
			metadata, err := dbapi.GetBatchProperties(opts.Name)
			if err != nil {
				sendErr = err
			}

			if sendErr != nil {
				resp := ClientMessage{
					ClientID:    msg.ClientID,
					MessageType: "server_error",
					Error:       fmt.Sprintf("%v", sendErr),
				}

				jsnMsg, err := json.Marshal(resp)
				if err != nil {
					log.Printf("failed to marshal struct into JSON : %v", err)
					return
				}

				conn.WriteMessage(websocket.TextMessage, jsnMsg)

			} else {
				res := BatchPayload{
					FetchOptions: opts,
					Metadata:     metadata,
					Batch:        batch,
				}
				batchPayload, err := json.Marshal(res)
				if err != nil {
					log.Printf("failed to marshal struct into JSON : %v", err)
					return
				}
				resp := ClientMessage{
					ClientID:    msg.ClientID,
					MessageType: "fetched_batch",
					Payload:     string(batchPayload),
				}

				jsnResp, err := json.Marshal(resp)
				if err != nil {
					log.Printf("failed to marshal struct into JSON : %v", err)
					return
				}
				conn.WriteMessage(websocket.TextMessage, jsnResp)
			}
		default:
			log.Printf("Unknown message type: %s", msg.MessageType)
		}

	}
}

func stats(clientID string) error {

	ws, ok := clients[clientID]
	if !ok {
		return fmt.Errorf("no websocket for client '%s'", clientID)
	}

	stats, err := dbapi.GetStats()
	if err != nil {
		return fmt.Errorf("faild to get stats from db : '%v'", err)
	}

	jsn, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("faild to marshal stats struct : '%v'", err)
	}

	err = ws.WriteMessage(websocket.TextMessage, jsn)
	if err != nil {

		log.Printf("Error: removing websocket '%s'\n", clientID)
		delete(clients, clientID)
		return fmt.Errorf("failed to write stats to websocket : %v", err)
	}

	return nil
}

func keepAlive() {
	msg := ClientMessage{
		MessageType: "keep_alive",
	}

	t := time.NewTicker(23 * time.Second)
	for _ = range t.C {

		clientMutex.RLock()
		for id, ws := range clients {
			msg.ClientID = id
			jsn, err := json.Marshal(msg)
			if err != nil {
				log.Fatalf("failed to marshal JSON : %v", err)
				return
			}

			err = ws.WriteMessage(websocket.TextMessage, jsn)
			if err != nil {
				log.Printf("websocket error for client ID '%s' : %v", id, err)
				ws.Close()
				delete(clients, id)
			}
		}

		clientMutex.RUnlock()
	}
}

func main() {

	host := flag.String("h", "localhost", "Server `host`")
	port := flag.String("p", "7337", "Server `port`")
	dbPath := flag.String("db", "test.db", "Sqlite3 DB path")

	// live db instance:
	err := dbapi.Open(*dbPath)
	if err != nil {
		log.Fatalf("failed to open db '%s' : %v\n", *dbPath, err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/ws/{client_id}", wsHandler)

	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static/"))))

	go keepAlive()

	r.StrictSlash(true)
	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", *host, *port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Server started on %s\n", srv.Addr)

	log.Fatal(srv.ListenAndServe())
	fmt.Println("No fun")

}
