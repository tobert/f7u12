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
  var serialize = function (game, dir, tile_idx) {
    var now = performance.now();

    var tile_val = 0;
    if (tile_idx != null) {
       tile_val = game.cells[tile_idx];
    }

    var out = {
      "grid_id":   game.uuid,
      "turn_id":   game.sequence,
      "offset_ms": now - game.started,
      "turn_ms":   now - game.last_turn,
      "player":    game.target, // temp, replace with text box
      "score":     game.score,
      "tile_val":  tile_val,
      "tile_idx":  tile_idx,
      "dir":       dir,
      "grid":      game.cells
    };

    console.log(out);

    return JSON.stringify(out);
  };

  var move = function (game, dir) {
    var changed = game.move(dir);
    var new_tile_idx = game.insert();

    if (!changed) {
      return;
    }

    $.ajax({
      url: "/grid",
      type: "PUT",
      dataType: "json",
      data: serialize(game, dir, new_tile_idx),
      success: function () { console.log("XHR Succeeded."); }
    });

    game.last_turn = performance.now();
  };

  var make_game = function(target) {
    var game = new F7U12(4); // 4x4 grid
        game.init(2); // start with 2 tiles
        game.render(target);
        game.make_dpad(target);

    // jquery swipe plugin
    $(target).swipe({ swipe: function(e, direction) { move(game, direction); } });

    // a simple dpad for testing with the mouse
    ["up", "down", "left", "right"].forEach(function (dir) {
      $(target + " #" + dir).on("click", function () { move(game, dir); });
    });

    // fill in some fields that are used for serializing the game
    game["score"] = 0;
    game["started"] = performance.now(); // high-res timestamp
    game["last_turn"] = performance.now();
    game["uuid"] = UUIDjs.create(1).toString();

    return game;
  };

  // disable default swipe events (e.g. ios overscroll)
  document.body.addEventListener('touchmove', function(event) { event.preventDefault(); }, false);

  // designed to have multiple players on the same touch screen
  // for now, just one player is configured
  [1].forEach(function (pnum) {
    var target = "#player" + pnum + "-container";
    var game = make_game(target);
    start_websocket(game.uuid);

    // send the starting board to the server
    $.ajax({
      url: "/grid",
      type: "PUT",
      dataType: "json",
      data: serialize(game, null, null),
      success: function () { console.log("Game init succeeded."); }
    });
  });
});
