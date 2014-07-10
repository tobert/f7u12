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
      dataType: "json",
      data: game.serialize(),
      success: function () { console.log("XHR Succeeded."); }
    });

    game.last_turn = performance.now();

    console.log(game.serialize());
  };

  // monkey patch this in so it can be run from the script in the HTML
  // TODO: find a cleaner way to expose this ...
  F7U12.run_ai_fun = function (fun) {
    d3.select("#player1-container").selectAll("div").remove();

    var game = new F7U12(4); // 4x4 grid
        game.init(2); // start with 2 tiles
        game.set_name("AI");
        game.uuid = UUIDjs.create(1).toString();
        game.render("#player1-container");
        start_websocket(game.uuid, "#log-container");

    // send the starting board to the server
    $.ajax({
      url: "/grid",
      type: "PUT",
      data: game.serialize(null, null),
      success: function () { console.log("Game init succeeded."); }
    });

    // wrap move() so it's visible in the eval
    game.mv = function (dir) { move(this, dir); };

    var iter = 0;
    while (!game.over()) {
      iter++;
      // calls fun() with this = game, arg0: game, arg1: iteration number
      fun.call(game, game, iter);
    }
  };
});
