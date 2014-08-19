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
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gorilla/mux"
)

type BBdata struct {
	MacAddress string    `json:"mac_address"`
	Timestamp  time.Time `json:"timestamp"`
	Direction  string    `json:"direction"`
	Count      int       `json:"count"`
	Period     int       `json:"period"`
	Weight     int       `json:"weight"`
	Min        int       `json:"min"`
	Max        int       `json:"max"`
	Sum        int       `json:"sum"`
	Mean       int       `json:"mean"`
	Variance   int       `json:"variance"`
	Stdev      int       `json:"stdev"`
	RFpcnt     int       `json:"rf_pcnt"`
	RRpcnt     int       `json:"rr_pcnt"`
	LFpcnt     int       `json:"lf_pcnt"`
	LRpcnt     int       `json:"lr_pcnt"`
	RFmean     int       `json:"rf_mean"`
	RRmean     int       `json:"rr_mean"`
	LFmean     int       `json:"lf_mean"`
	LRmean     int       `json:"lr_mean"`
	RFstdev    int       `json:"rf_stdev"`
	RRstdev    int       `json:"rr_stdev"`
	LFstdev    int       `json:"lf_stdev"`
	LRstdev    int       `json:"lr_stdev"`
}

func BBoardHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mac := vars["mac"]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("Websocket upgrade failed: %s\n", err)
	}
	defer conn.Close()

	query := `
SELECT mac, ts, direction,
 count, period, weight, min, max, sum, mean, variance, stdev,
 rf_pcnt, rr_pcnt, lf_pcnt, lr_pcnt,
 rf_mean, rr_mean, lf_mean, lr_mean,
 rf_stdev, rr_stdev, lf_stdev, lr_stdev
FROM balance_board_input
WHERE bucket=? AND mac=?
LIMIT 1
`
	q := cass.Query(query) // prepared statement
	prev := time.Now()

	go func() {
		ticker := time.NewTicker(time.Millisecond * 10)
		for now := range ticker.C {
			bucket := now.Truncate(time.Minute)
			q.Bind(&bucket, &mac)

			bbd := BBdata{}
			err := q.Scan(
				&bbd.MacAddress, &bbd.Timestamp, &bbd.Direction,
				&bbd.Count, &bbd.Period, &bbd.Weight, &bbd.Min, &bbd.Max, &bbd.Sum, &bbd.Mean, &bbd.Variance, &bbd.Stdev,
				&bbd.RFpcnt, &bbd.RRpcnt, &bbd.LFpcnt, &bbd.LRpcnt,
				&bbd.RFmean, &bbd.RRmean, &bbd.LFmean, &bbd.LRmean,
				&bbd.RFstdev, &bbd.RRstdev, &bbd.LFstdev, &bbd.LRstdev,
			)
			// plenty of errors on no data in current bucket, ignore for now
			// TODO: check error type and log on errors other than NotFound

			// don't send on the websocket if it's the same read as last time around
			if bbd.Timestamp == prev {
				continue
			}
			prev = bbd.Timestamp

			js, err := json.Marshal(bbd)
			if err != nil {
				log.Printf("JSON marshal failed: %s\n", err)
			}

			err = conn.WriteMessage(websocket.TextMessage, js)
			if err != nil {
				log.Printf("Failed to send %d bytes on websocket: %s", len(js), err)
				return
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
