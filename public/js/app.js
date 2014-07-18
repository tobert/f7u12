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
      dataType: "json",
      data: game.serialize(),
      success: function () { console.log("XHR Succeeded."); }
    });

    game.score_div.text("Score: " + game.score);

    game.last_turn = performance.now();
  };

  var make_game = function(target) {
     var score_div = d3.select(target).append("div")
       .attr("class", "f7u12-score")
       .text("Score: 0");

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

    document.onkeydown = function() {
      switch (window.event.keyCode) {
        case 37: move(game, "left"); break;
        case 38: move(game, "up"); break;
        case 39: move(game, "right"); break;
        case 40: move(game, "down"); break;
      }
    };

    game["score"] = 0;
    game["uuid"] = UUIDjs.create(1).toString();
    game["score_div"] = score_div;

    return game;
  };

  // disable default swipe events (e.g. ios overscroll)
  document.body.addEventListener('touchmove', function(event) { event.preventDefault(); }, false);

  // designed to have multiple players on the same touch screen
  // for now, just one player is configured
  [1].forEach(function (pnum) {
    var target = "#player" + pnum + "-container";
    var game = make_game(target);
        game.set_name("player" + pnum);
    start_websocket(game.uuid, target);

    // send the starting board to the server
    $.ajax({
      url: "/grid",
      type: "PUT",
      dataType: "json",
      data: game.serialize(null, null),
      success: function () { console.log("Game init succeeded."); }
    });
  });
});
