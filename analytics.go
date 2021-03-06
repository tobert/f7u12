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

type AvgScoreByTurn struct {
	Name     string  `json:"name"` // e.g. human, ai
	TurnId   int     `json:"turn_id"`
	AvgScore float32 `json:"avg_score"`
}
type AvgScoreByTurns []AvgScoreByTurn

// TODO: if this ever gets adapted for use on a public site
// the list of games will get crazy so it'll need a limit
// but for now, there are only a small handful so keep it simple
func GetRecentGames(cass *gocql.Session) (games PlayerGames, err error) {
	games = make(PlayerGames, 0)

	// player_game has a reverse comparator, so grab everything and
	// throw away duplicate players here
	iq := cass.Query(`SELECT player, game_id FROM player_game`).Iter()

	seen := make(map[string]bool)
	for {
		pg := PlayerGame{}
		ok := iq.Scan(&pg.Player, &pg.GameId)
		if ok {
			// only get the first record for each player name
			if _, ok := seen[pg.Player]; !ok {
				games = append(games, pg)
				seen[pg.Player] = true
			}
		} else {
			break
		}
	}
	if err = iq.Close(); err != nil {
		log.Printf("Error during query (GetRecentGames): %s", err)
	}

	return games, err
}

// GET /recent
func RecentGamesHandler(w http.ResponseWriter, r *http.Request) {
	games, err := GetRecentGames(cass)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cassandra query failed: %s", err), 500)
		return
	}

	json, err := json.Marshal(games)
	if err != nil {
		log.Printf("JSON marshal failed: %s\n", err)
		http.Error(w, fmt.Sprintf("Marshaling JSON failed: %s", err), 500)
	}

	w.Write(json)
}

func GetTopGames(cass *gocql.Session, dimension string) (tgs TopGames, err error) {
	tgs = make(TopGames, 0)

	query := `SELECT dimension, rank, game_id, score FROM top_games WHERE dimension=?`
	iq := cass.Query(query, dimension).Iter()

	for {
		tg := TopGame{}
		ok := iq.Scan(&tg.Dimension, &tg.Rank, &tg.GameId, &tg.Score)
		if ok {
			tgs = append(tgs, tg)
		} else {
			break
		}
	}
	if err = iq.Close(); err != nil {
		log.Printf("Error during query (GetTopGames): %s", err)
	}

	return
}

// GET /top_games/ai_topN
func TopGamesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tgs, err := GetTopGames(cass, vars["dimension"])
	if err != nil {
		http.Error(w, fmt.Sprintf("Cassandra query failed: %s", err), 500)
		return
	}

	json, err := json.Marshal(tgs)
	if err != nil {
		log.Printf("JSON marshal failed: %s\n", err)
		http.Error(w, fmt.Sprintf("Marshaling JSON failed: %s", err), 500)
	}

	w.Write(json)
}

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
		log.Printf("Error during query (GetCounts): %s", err)
	}

	return counts, err
}

// GET /counts
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

func GetDirCounts(cass *gocql.Session, game_id gocql.UUID) (dcs DirCounts, err error) {
	dcs = make(DirCounts, 0)

	query := `SELECT game_id, counts FROM dir_counts WHERE game_id=?`
	iq := cass.Query(query, game_id).Iter()

	for {
		dc := DirCount{}
		ok := iq.Scan(&dc.GameId, &dc.Counts)
		if ok {
			dcs = append(dcs, dc)
		} else {
			break
		}
	}
	if err = iq.Close(); err != nil {
		log.Printf("Error during query (GetDirCounts): %s", err)
	}

	return
}

// GET /dir_counts/<uuid>
func DirCountsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	game_id, err := gocql.ParseUUID(vars["game_id"])
	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse game_id (uuid expected): '%s'", err), 500)
		return
	}

	dcs, err := GetDirCounts(cass, game_id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cassandra query failed: %s", err), 500)
		return
	}

	json, err := json.Marshal(dcs)
	if err != nil {
		log.Printf("JSON marshal failed: %s\n", err)
		http.Error(w, fmt.Sprintf("Marshaling JSON failed: %s", err), 500)
	}

	w.Write(json)
}

func GetAvgScoreByTurn(cass *gocql.Session) (avgs AvgScoreByTurns, err error) {
	// allocate an osbscenely large list so values can be injected by index
	// which is used in the websocket handler to access by turn id
	avgs = make(AvgScoreByTurns, 2048)

	iq := cass.Query(`SELECT name, turn_id, avg_score FROM avg_score_by_turn`).Iter()

	for {
		avg := AvgScoreByTurn{}
		ok := iq.Scan(&avg.Name, &avg.TurnId, &avg.AvgScore)
		if ok {
			avgs[avg.TurnId] = avg
		} else {
			break
		}
	}
	if err = iq.Close(); err != nil {
		log.Printf("Error during query (GetAvgScoreBy): %s", err)
	}

	return avgs, err
}

// GET /avg_score_by_turn
func AvgScoreByTurnHandler(w http.ResponseWriter, r *http.Request) {
	avgs, err := GetAvgScoreByTurn(cass)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cassandra query failed: %s", err), 500)
		return
	}

	json, err := json.Marshal(avgs)
	if err != nil {
		log.Printf("JSON marshal failed: %s\n", err)
		http.Error(w, fmt.Sprintf("Marshaling JSON failed: %s", err), 500)
	}

	w.Write(json)
}
