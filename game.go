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

type Grid struct {
	GridId    gocql.UUID `json:"grid_id"`   // timeuuid
	TurnId    int        `json:"turn_id"`   // turn number
	OffsetMs  float64    `json:"offset_ms"` // time offset from beginning of game (ms)
	TurnMs    float64    `json:"turn_ms"`   // time elapsed between turns
	Player    string     `json:"player"`    // arbitrary string, player name
	Score     float32    `json:"score"`     // score at the end of the turn
	TileVal   int        `json:"tile_val"`  // the new tile value put on the board
	TileIdx   int        `json:"tile_idx"`  // index on the grid where the new tile was placed
	Direction string     `json:"dir"`       // up down left right init
	Grid      []int      `json:"grid"`      // every value in the grid
}

type Grids []Grid

func (g *Grid) Save(cass *gocql.Session) error {
	query := `INSERT INTO grids (grid_id, turn_id, offset_ms, turn_ms, player, score, tile_val, tile_idx, direction, grid) VALUES (?,?,?,?,?,?,?,?,?,?)`
	return cass.Query(query, g.GridId, g.TurnId, g.OffsetMs, g.TurnMs, g.Player, g.Score, g.TileVal, g.TileIdx, g.Direction, g.Grid).Exec()
}

func GetGrids(cass *gocql.Session, id gocql.UUID, min_turn_id, max_turn_id int) (list Grids, err error) {
	list = make(Grids, 0)

	query := `SELECT grid_id, turn_id, offset_ms, turn_ms, player, score, tile_val, tile_idx, direction, grid FROM grids WHERE grid_id=?`
	iq := cass.Query(query, id).Iter()

	for {
		g := Grid{}
		ok := iq.Scan(&g.GridId, &g.TurnId, &g.OffsetMs, &g.TurnMs, &g.Player, &g.Score, &g.TileVal, &g.TileIdx, &g.Direction, &g.Grid)
		if ok {
			list = append(list, g)
		} else {
			break
		}
	}
	if err = iq.Close(); err != nil {
		log.Printf("Error during Cassandra query: %s", err)
	}

	return list, err
}

func GridHandler(w http.ResponseWriter, r *http.Request) {
	g := Grid{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&g)
	if err != nil {
		log.Printf("PUT invalid json data: %s", err)
		http.Error(w, fmt.Sprintf("PUT invalid json data: %s", err), 500)
	}

	switch r.Method {
	case "PUT":
		err := g.Save(cass)
		if err != nil {
			log.Printf("Write to Cassandra failed: %s", err)
			http.Error(w, "Write to Cassandra failed!", 500)
		}
	default:
		http.Error(w, fmt.Sprintf("method '%s' not implemented", r.Method), 500)
		return
	}
}
