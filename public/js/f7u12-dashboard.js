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
 * F7U12.Dash: a dashboard for the F7U12 game
 */

// create a leaderboard widget
F7U12.Leaderboard = function (target, title) {
  var lb = this;

  lb.table = d3.select(target).append("table")
    .classed("f7u12-leaderboard", true) // see ../css/dashboard.css
    .classed("table", true) // bootstrap
    .classed("table-condensed", true); // bootstrap

  lb.table.append("caption").text(title);

  lb.tbody = lb.table.append("tbody");
};

F7U12.Leaderboard.prototype.render = function (data) {
  var lb = this;

  lb.rows = lb.tbody.selectAll("tr")
    .data(data)
    .enter()
    .append("tr");

  // rank game_id score
  lb.cells = lb.rows.selectAll("td")
    .data(function (d) { return [d.game_id, d.score]; })
    .enter()
    .append("td")
    .text(function (d) { return d; });
};

F7U12.Leaderboard.prototype.update = function (data) {
  var lb = this;
  lb.cells.data(data, function (d) { return [d.game_id, d.score]; });
  lb.cells.exit().remove();
  lb.cells.enter()
    .append("td")
    .text(function (d) { return d; });
};

// vim: et ts=2 sw=2 ai smarttab
