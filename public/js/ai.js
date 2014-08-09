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
 *
 * requires: f7u12, jQuery + swipe plugin, D3 3.x
 * supports: Chromium/Chrome, others may work but are not tested
 * see also:
 *   https://developer.mozilla.org/en-US/docs/Web/API/Performance.now()
 */

$(function() {
  var move = function (game, dir) {
    var changed = game.move(dir);
    if (!changed) {
      return;
    }

    var new_tile_idx = game.insert();
    $.ajax({
      url: "/grid",
      type: "PUT",
      data: game.serialize(),
      success: function () { console.log("XHR Succeeded."); }
    });

    game.last_turn = performance.now();
  };

  var naive_mover = function (game) {
    if (!game.can_move("up")) {
      // when both left and right are possible, choose the one that merges the most cells
      if (game.can_move("right") && game.can_move("left")) {
        var left_up = game.clone();
            left_up.move("left");
            left_up.move("up");

        var right_up = game.clone();
            right_up.move("right");
            right_up.move("up");

        if (left_up.empty_cells() > right_up.empty_cells()) {
          move(game, "left");
        } else {
          move(game, "right");
        }
      // try left
      } else if (move(game, "left")) {
        return;
      // try right
      } else if (move(game, "right")) {
        return;
      // when all else fails, down
      } else {
        move(game, "down");
      }
    // always move up when you can
    } else {
      move(game, "up");
    }
  };

  // export for loops, etc.
  F7U12.run = function () {
    var game = new F7U12(4); // 4x4 grid
        game.init(2); // start with 2 tiles
        game.name = "AI";
        game.uuid = UUIDjs.create(1).toString();

    // send the starting board to the server
    $.ajax({
      url: "/grid",
      type: "PUT",
      data: game.serialize(null, null),
      success: function () { console.log("Game init succeeded."); }
    });

    while (!game.over()) {
      naive_mover(game);
    }

    return game;
  };
});
