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
	"fmt"
	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := gocql.ParseUUID(vars["grid_id"])
	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse grid_id (uuid expected): '%s'", err), 500)
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("Websocket upgrade failed: %s\n", err)
	}
	defer conn.Close()

	// TODO: this is a stub
	go func() {
		var js = []byte(fmt.Sprintf("{\"grid_id\": \"%s\"}\n", id))
		for {
			select {
			case <-time.After(time.Second):
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
