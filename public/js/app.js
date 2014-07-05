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
  // create player 1's game board
  var p1 = new F7U12(4, "#player1-game");
      p1.init(2);

  // create player 2's game board
  var p2 = new F7U12(4, "#player2-game");
      p2.init(2);

  // disable default swipe events (e.g. ios overscroll)
  document.body.addEventListener('touchmove', function(event) { event.preventDefault(); }, false);

  // wire the jquery swipe plugin for player 1
  $("#player1-game").swipe({
    swipe: function(event, direction, distance, duration, fingerCount, fingerData) {
      console.log("player 1 swiped ", direction);
      p1.move(direction);
    }
  });

  $("#player1-dpad #up"   ).on("click", function (e) { p1.move("up"); });
  $("#player1-dpad #down" ).on("click", function (e) { p1.move("down"); });
  $("#player1-dpad #left" ).on("click", function (e) { p1.move("left"); });
  $("#player1-dpad #right").on("click", function (e) { p1.move("right"); });

  // wire the jquery swipe plugin for player 2
  $("#player2-game").swipe({
    swipe: function(event, direction, distance, duration, fingerCount, fingerData) {
      console.log("player 2 swiped ", direction);
      p2.move(direction);
    }
  });

  $("#player2-dpad #up"   ).on("click", function (e) { p2.move("up"); });
  $("#player2-dpad #down" ).on("click", function (e) { p2.move("down"); });
  $("#player2-dpad #left" ).on("click", function (e) { p2.move("left"); });
  $("#player2-dpad #right").on("click", function (e) { p2.move("right"); });
});
