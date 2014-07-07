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
	"log"
	"net/http"
)

func jsonOut(w http.ResponseWriter, r *http.Request, data interface{}) {
	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(js)
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
		err := g.Save(cass);
		if err != nil {
			log.Printf("Write to Cassandra failed: %s", err);
			http.Error(w, "Write to Cassandra failed!", 500)
		}
	default:
		http.Error(w, fmt.Sprintf("method '%s' not implemented", r.Method), 500)
		return
	}

	jsonOut(w, r, g)
}
