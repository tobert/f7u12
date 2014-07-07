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
	"net/http"
)

var cass *gocql.Session

func main() {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "f7u12"
	cluster.Consistency = gocql.Quorum

	var err error
	cass, err = cluster.CreateSession()
	if err != nil {
		panic(fmt.Sprintf("Error creating Cassandra session: %v", err))
	}
	defer cass.Close()

	http.HandleFunc("/grid", GridHandler)

	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./public/js/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./public/css/"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/index.html")
	})

	http.ListenAndServe(":8080", nil)
}
