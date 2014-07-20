fffffffuuuuuuuuuuuu
===================

This is a demo project for generating metrics to store in
Cassandra and query with CQL or Spark.

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
game will run fine on its own in a normal webserver. The XHRs
will fail but it does not break the game. It does reference
a couple CDN links. The assets are also checked in, so making it
work offline only requires changing the CDN links in the html files.

The Spark job needs Cassandra running.

Components
==========

* Javascript 2048 game widget called f7u12
* Go backend for storing and reading metrics in Cassandra
* Scala job that analyzes all game stats in Cassandra
* Bootstrap 3 / D3.js dashboard(s) to display metrics

Status
======

* Game is playable and sends every move to the Go app immediately.
* Go backend is basically complete. Moving to a more REST-compliant URL scheme might be nice.
* Spark job processes the move data and stores aggregates in Cassandra (that are readable through the Go app).
* Dashboards in progress.






Security
========

This application is intended as a demo and does not always follow best security practices.

Dependencies
============

Dependencies included in this repo have their own licenses.

* D3
* jQuery TouchSwipe Plugin https://github.com/mattbryson/TouchSwipe-Jquery-Plugin
* uuid.js https://github.com/pnegri/uuid-js

