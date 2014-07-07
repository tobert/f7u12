CREATE KEYSPACE f7u12 WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 3 };

use f7u12;

CREATE TABLE grids (
    grid_id   timeuuid,
	turn_id   int,
    offset_ms double,
    turn_ms   double,
    player    text,
    score     float,
    tile_val  int,
	tile_idx  int,
    direction text,
    grid      list<int>,
    PRIMARY KEY(grid_id, turn_id)
);

CREATE TABLE player_grid (
    player  text,
    grid_id timeuuid,
    PRIMARY KEY(player, grid_id)
);
