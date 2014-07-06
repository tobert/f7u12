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
	"github.com/gocql/gocql"
)

type Grid struct {
	GridId    gocql.UUID `json:"grid_id"`
	OffsetMs  int            `json:"offset_ms"` // time offset from beginning of game (ms)
	Player    string         `json:"player"`    // arbitrary string
	Score     float64        `json:"score"`
	NewTile   int            `json:"new_tile"`
	Direction string         `json:"dir"`
	Grid      []int          `json:"grid"`
}

func (g *Grid) Save(cass *gocql.Session) error {
	query := `INSERT INTO grids (grid_id, offset_ms, player, score, new_tile, direction, grid) VALUES (?,?,?,?,?,?,?)`
	return cass.Query(query, g.GridId, g.OffsetMs, g.Player, g.Score, g.NewTile, g.Direction, g.Grid).Exec()
}
