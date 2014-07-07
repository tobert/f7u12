fffffffuuuuuuuuuuuu
===================

This is a demo project for generating metrics to store in
Cassandra and query with CQL or Spark.

Usage
=====

Build & start the server then browse to http://localhost:8080

```
go get -u github.com/gocql/gocql
go build
./f7u12
```

Components
==========

* simple 2048 widget called f7u12
* 2-player page with touch controls
    * TODO: add player name entry
* Go backend for storing metrics in Cassandra
* Schema for storing metrics and game history

Status
======

The game is playable. It isn't wired up to anything and
needs some CSS love, but the mechanics are there and ready to start
tracking.

Security
========

This application is intended as a demo and does not always follow best security practices.

Dependencies
============

Dependencies included in this repo have their own licenses.

* D3
* jQuery TouchSwipe Plugin https://github.com/mattbryson/TouchSwipe-Jquery-Plugin
* uuid.js https://github.com/pnegri/uuid-js

