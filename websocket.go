package main

/*
 * Copyright 2014 Albert P. Tobey <atobey@datastax.com> @AlTobey
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import (
	"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"time"
)

type GameScoreData struct {
	AvgScoreByTurn
	Score float32 `json:"score"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	game_id, err := gocql.ParseUUID(vars["game_id"])
	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse game_id (uuid expected): '%s'", err), 500)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("Websocket upgrade failed: %s\n", err)
	}
	defer conn.Close()

	go func() {
		for {
			select {
			case <-time.After(time.Second):
				game, err := GetGame(cass, game_id)
				if err != nil {
					log.Printf("Cassandra query failed to get Game data: %s\n", err)
				}

				avgs, err := GetAvgScoreByTurn(cass)
				if err != nil {
					http.Error(w, fmt.Sprintf("Cassandra query failed to get avg data: %s", err), 500)
					return
				}

				// possibly fragile once the spark job writes AI into the
				// avg score table, but for now it's fine
				gsds := make([]GameScoreData, len(game))
				for i, grid := range game {
					gsds[i].TurnId = grid.TurnId
					gsds[i].Score = grid.Score
					gsds[i].Name = avgs[grid.TurnId].Name
					gsds[i].AvgScore = avgs[grid.TurnId].AvgScore
				}

				js, err := json.Marshal(gsds)
				if err != nil {
					log.Printf("JSON marshal failed: %s\n", err)
				}

				err = conn.WriteMessage(websocket.TextMessage, js)
				if err != nil {
					log.Printf("Failed to send %d bytes on websocket: %s", len(js), err)
					return
				}
			}
		}
	}()

	// TODO: this is a stub
	for {
		mt, payload, err := conn.ReadMessage()
		if err != nil {
			if err != io.EOF {
				log.Printf("conn.ReadMessage failed: %s\n", err)
				return
			}
		}

		switch mt {
		case websocket.BinaryMessage:
			log.Printf("Ignoring binary message: %q\n", payload)
		case websocket.TextMessage:
			log.Printf("Ignoring text message: %q\n", payload)
		default:
			log.Printf("Invalid message type %d\n", mt)
			return
		}
	}
}
