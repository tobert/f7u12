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
	"log"
	"net/http"
)

// stored in cassandra time ordered so it's easy to find the
// most recent/current game, updated in game.go on every write
type PlayerGame struct {
	Player string     `json:"player"`  // arbitrary string, player name
	GameId gocql.UUID `json:"game_id"` // timeuuid
}
type PlayerGames []PlayerGame

// topN game scores, e.g. ai_topN, human_topN
// computed with Spark and written to Cassandra
type TopGame struct {
	Dimension string     `json:"dimension"` // e.g. ai_topN
	Rank      int        `json:"rank"`      // 1-N (no 0)
	GameId    gocql.UUID `json:"game_id"`   // timeuuid
	Score     float32    `json:"score"`     // score at the end of the turn
}
type TopGames []TopGame

// various simple counters, computed with Spark and stored in Cassandra
type Count struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}
type Counts []Count

// global count of direction choices
// e.g. DirCount{uuid, "up", 9876}
type DirCount struct {
	GameId gocql.UUID     `json:"game_id"` // timeuuid
	Counts map[string]int `json:"counts"`  // {"up" => 999, "down" => 123, ...}
}
type DirCounts []DirCount

func GetCounts(cass *gocql.Session) (counts Counts, err error) {
	counts = make(Counts, 0)

	iq := cass.Query(`SELECT name,value FROM counts`).Iter()

	for {
		c := Count{}
		ok := iq.Scan(&c.Name, &c.Value)
		if ok {
			counts = append(counts, c)
		} else {
			break
		}
	}
	if err = iq.Close(); err != nil {
		log.Printf("Error during Cassandra query: %s", err)
	}

	return counts, err
}

func CountsHandler(w http.ResponseWriter, r *http.Request) {
	counts, err := GetCounts(cass)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cassandra query failed: %s", err), 500)
		return
	}

	json, err := json.Marshal(counts)
	if err != nil {
		log.Printf("JSON marshal failed: %s\n", err)
		http.Error(w, fmt.Sprintf("Marshaling JSON failed: %s", err), 500)
	}

	w.Write(json)
}
