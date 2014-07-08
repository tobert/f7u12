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
function start_websocket(grid_id) {
  var sock = new WebSocket("ws://" + window.location.host + "/ws/" + grid_id);
  var lines = new Array(128);

  sock.onerror = function (e) {
    console.log("socket error", e);
  };

  sock.onopen = function (e) {
    var panel = d3.select("#player1-container")
      .append("div")
      .attr("class", "ws-panel")
      .attr("id", "player1-ws-panel");

    var ta = panel.append("textarea")
      .attr("disabled", true)
      .attr("cols", 120)
      .attr("rows", 8);

    sock.onmessage = function(msg) {
      console.log("Message: ", msg.data);

      lines.push(msg.data.replace(/[\r\n]+/, " "));
      if (lines.length == 128) {
        lines.shift();
      }

      ta.each(function () {
        this.value = lines.join("\r\n");
        // TODO: needs a pause indicator so users can scroll back
        // maybe make it automatic on scrollback ...
        this.scrollTop = this.scrollHeight;
      });
    };
  };
}
