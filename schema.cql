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
    counts  map<text,int>,
    PRIMARY KEY(game_id)
);

CREATE TABLE avg_score_by_turn (
    name      text,
    turn_id   int,
    avg_score float,
    PRIMARY KEY((name,turn_id), avg_score)
);

CREATE TABLE balance_board_input (
    bucket      timestamp,
    mac         text,
    ts          timestamp,
    count       int,
    period      int,
    weight      int,
    min         int,
    max         int,
    sum         int,
    mean        int,
    variance    int,
    stdev       int,
    rf_pcnt     int,
    rr_pcnt     int,
    lf_pcnt     int,
    lr_pcnt     int,
    rf_mean     int,
    rr_mean     int,
    lf_mean     int,
    lr_mean     int,
    rf_stdev    int,
    rr_stdev    int,
    lf_stdev    int,
    lr_stdev    int,
    PRIMARY KEY((bucket,mac), ts)
);

// vim: et ts=4 sw=4 ai smarttab
