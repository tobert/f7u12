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
	"flag"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var addrFlag, cqlFlag, ksFlag string
var cass *gocql.Session

func init() {
	flag.StringVar(&addrFlag, "addr", ":9000", "IP:PORT or :PORT address to listen on")
	flag.StringVar(&cqlFlag, "cql", "127.0.0.1", "IP or IP:port of the Cassandra CQL service")
	flag.StringVar(&ksFlag, "ks", "f7u12", "keyspace containing the f7u12 schema")
}

func main() {
	flag.Parse()

	cluster := gocql.NewCluster(cqlFlag)
	cluster.Keyspace = ksFlag
	cluster.Consistency = gocql.One

	var err error
	cass, err = cluster.CreateSession()
	if err != nil {
		panic(fmt.Sprintf("Error creating Cassandra session: %v", err))
	}
	defer cass.Close()

	r := mux.NewRouter()
	r.HandleFunc("/grid", GridHandler)
	r.HandleFunc("/game/{game_id:[-a-f0-9]+}", GameHandler)
	r.HandleFunc("/counts", CountsHandler)
	r.HandleFunc("/recent", RecentGamesHandler)
	r.HandleFunc("/avg_score_by_turn", AvgScoreByTurnHandler)
	r.HandleFunc("/top_games/{dimension:[-_a-zA-Z0-9]+}", TopGamesHandler)
	r.HandleFunc("/dir_counts/{game_id:[-a-f0-9]+}", DirCountsHandler)
	r.HandleFunc("/ws/{game_id:[-a-f0-9]+}", WsHandler)
	r.HandleFunc("/balance_board/{mac:[:0-9a-f]+}", BBoardHandler)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	http.Handle("/", r)
	err = http.ListenAndServe(addrFlag, nil)
	if err != nil {
		log.Fatalf("net.http could not listen on address '%s': %s\n", addrFlag, err)
	}
}
