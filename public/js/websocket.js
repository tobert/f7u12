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
  var not_initialized = true;

  var score_chart = function (data) {
    var svg = dimple.newSvg("#score-chart", 200, 120);
    var chart = new dimple.chart(svg, data);
    chart.setBounds(20, 20, 160, 80);
    var x = chart.addMeasureAxis("x", "offset_ms");
        x.hidden = true;
    var y = chart.addMeasureAxis("y", "score");
        y.hidden = true;
    var s = chart.addSeries("score", dimple.plot.line);
        s.lineWeight = 1;
    chart.draw();

    chart.update = function (data) {
      chart.data = data;
      chart.draw();
    };

    return chart;
  };

  sock.onerror = function (e) {
    console.log("socket error", e);
  };

  sock.onopen = function (e) {
    var chart;
    sock.onmessage = function(msg) {
      var data = JSON.parse(msg.data);
      if (not_initialized) {
        chart = score_chart(data);
        not_initialized = false;
      } else {
        chart.update(data);
      }
    };
  };
}
