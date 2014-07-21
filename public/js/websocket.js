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

// break with style for now - get it working ...
function start_websocket(grid_id, target) {
  var sock = new WebSocket("ws://" + window.location.host + "/ws/" + grid_id);

  sock.onerror = function (e) {
    console.log("socket error", e);
  };

  sock.onopen = function (e) {
    var chart;
    sock.onmessage = function(msg) {
      var data = JSON.parse(msg.data);
      var last = data.length - 1;
      var current_turn = parseInt(d3.select("#turn-id-value").text());
      console.log("TURN ID", current_turn);
      d3.select("#avg-score-value").text(parseInt(data[last].avg_score));
    };
  };
}
