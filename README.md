f7u12
=====

This is a demo project for generating metrics to store in
Cassandra and query with CQL and Spark.

One of the other open source implementations of 2048 says "The best game you'll ever hate." That made me think of rage guy, so here we are.

Usage
=====

Build & start the server then browse to http://localhost:9000

```
go get -u github.com/gocql/gocql
go get -u github.com/gorilla/mux
go get -u github.com/gorilla/websocket
go build
./f7u12 -help
Usage of ./f7u12:
  -addr=":9000": IP:PORT or :PORT address to listen on
  -cql="127.0.0.1": IP or IP:port of the Cassandra CQL service
  -ks="f7u12": keyspace containing the f7u12 schema
```

Dependencies
============

Only Cassandra has to be running to run and play the game. The
game uses the embedded webserver to serve assets. All UI is static
or generated client-side.

Without the Go app running, the XHRs
will fail but it does not break the game. Libraries are checked in
and referenced locally because conference networks are unreliable.

The Spark job needs Cassandra running and does not contact the Go app.

Docker
======

The Dockerfile included in this repo uses scratch as the base. The
f7u12 app needs to be built statically to work without libc.

Build with CGO\_ENABLED=0 to make sure you get a static binary.
`CGO_ENABLED=0 go build -a`

Components
==========

* Javascript 2048 game widget called f7u12
* Go backend for storing and reading metrics in Cassandra
* Scala job that analyzes all game stats in Cassandra
* Bootstrap 3 / D3.js dashboard(s) to display metrics

Status
======

* Game is playable and sends every move to the Go app immediately.
* Go backend is complete. Moving to a more REST-compliant URL scheme might be nice.
* Spark job processes the move data and stores aggregates in Cassandra (that are readable through the Go app).
* Dashboard was demoed. Lots of ways to expand it.

Security
========

This application is intended as a demo and does not always follow best security practices.

Dependencies
============

Dependencies included in this repo have their own licenses.

* D3
* jQuery TouchSwipe Plugin https://github.com/mattbryson/TouchSwipe-Jquery-Plugin
* uuid.js https://github.com/pnegri/uuid-js

