CREATE KEYSPACE f7u12 WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 3 };

use f7u12;

CREATE TABLE grid (
    grid_id   timeuuid,
    offset_ms int,
    player    text,
    score     float,
    new_tile  int,
    direction text,
    grid      list<int>,
    PRIMARY KEY(grid_id, offset_ms)
);

CREATE TABLE player_grid (
    player  text,
    grid_id timeuuid,
    PRIMARY KEY(player, grid_id)
);
