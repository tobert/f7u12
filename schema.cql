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
);
