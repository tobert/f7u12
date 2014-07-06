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
 */

$(function() {
  // disable default swipe events (e.g. ios overscroll)
  document.body.addEventListener('touchmove', function(event) { event.preventDefault(); }, false);

  var move = function (p, dir) {
      var data = p.move(dir);
      console.log("Move Data: ", data);
  };

  // make two grids with controls
  [1, 2].forEach(function (pnum) {
    var target = "#player" + pnum + "-container";
    var player = new F7U12(4, target); // 4x4 grid
    player.init(2); // start with 2 tiles

    // jquery swipe plugin
    $(target).swipe({ swipe: function(e, direction) { move(player, direction); } });

    // a simple dpad for testing with the mouse
    ["up", "down", "left", "right"].forEach(function (dir) {
      $(target + " #" + dir).on("click", function () { move(player, dir); });
    });

  });
});
