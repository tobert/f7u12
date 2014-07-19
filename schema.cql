CREATE KEYSPACE f7u12 WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 3 };

use f7u12;

CREATE TABLE grids (
    game_id   timeuuid,
    turn_id   int,
    offset_ms double,
    turn_ms   double,
    player    text,
    score     float,
    tile_val  int,
    tile_idx  int,
    direction text,
    state     list<int>,
    PRIMARY KEY(game_id, turn_id)
);

CREATE TABLE player_game (
    player  text,
    game_id timeuuid,
    PRIMARY KEY(player, game_id)
) WITH CLUSTERING ORDER BY (game_id DESC);

CREATE TABLE top_games (
	dimension text,
	rank      int,
	game_id   timeuuid,
	score     float,
	PRIMARY KEY(dimension, rank)
);

CREATE TABLE counts (
	name text,
	value int,
	PRIMARY KEY(name)
);

CREATE TABLE dir_counts (
    game_id timeuuid,
	dir     text,
	value   int,
	PRIMARY KEY(game_id, dir)
);

CREATE TABLE game_dir_counts (
	game_id timeuuid,
	counts  map<string,int>,
	PRIMARY KEY(game_id)
);

CREATE TABLE move_ms_histogram (
	game_id   timeuuid,
	bucket    int,
	turn_ms   double,
	PRIMARY KEY(game_id, bucket)
);

