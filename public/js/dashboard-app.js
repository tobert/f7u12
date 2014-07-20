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
  // one-time setup, creating elements, etc.
  document.body.addEventListener('touchmove', function(event) { event.preventDefault(); }, false);

  var game = new F7U12(4);
      game.render("#grid1");

  //var game = new F7U12(4);
  //    game.render("#grid2");

  var human_leaderboard = new F7U12.Leaderboard("#human_topN_leaderboard", "Human Top 10");
  var ai_leaderboard = new F7U12.Leaderboard("#ai_topN_leaderboard", "AI Top 10");

  var svg1 = dimple.newSvg("#graph1", 590, 400);
  var svg2 = dimple.newSvg("#graph2", 590, 400);
  var svg3 = dimple.newSvg("#graph3", 590, 400);

  // this is a little cheezy but demonstrates a working graph with dimple
  d3.json("/counts", function (data) {
    var myChart = new dimple.chart(svg3, data);
    myChart.setBounds(75, 30, 480, 330)
    myChart.addMeasureAxis("x", "value");
    var y = myChart.addCategoryAxis("y", "name");
    myChart.addSeries("value", dimple.plot.bar);
    myChart.draw();
  });

  d3.json("/counts", function (data) {
    var re = /^(ai|human)_(up|down|left|right)/;
    var ai_re = /^ai_(up|down|left|right)/;
    var human_re = /^human_(up|down|left|right)/;

    var ai_moves = data.filter(function (d) { return ai_re.test(d.name); })
      .map(function (d) { return { name: d.name.replace(re, "$2"), value: d.value }; });

    var human_moves = data.filter(function (d) { return human_re.test(d.name); })
      .map(function (d) { return { name: d.name.replace(re, "$2"), value: d.value }; });

    var ai_move_chart = new dimple.chart(svg2, ai_moves);
    ai_move_chart.setBounds(60, 30, 430, 330);
    ai_move_chart.addCategoryAxis("x", "name");
    ai_move_chart.addMeasureAxis("y", "value");
    ai_move_chart.addSeries("value", dimple.plot.bar);
    ai_move_chart.draw();

    var human_move_chart = new dimple.chart(svg1, human_moves);
    human_move_chart.setBounds(60, 30, 430, 330);
    human_move_chart.addCategoryAxis("x", "name");
    human_move_chart.addMeasureAxis("y", "value");
    human_move_chart.addSeries("value", dimple.plot.bar);
    human_move_chart.draw();
  });

  var update = function (url, widget) {
    $.ajax({
      url: url,
      dataType: "json",
      success: function (data) {
        //console.log(url, data);
        widget.render(data);
      }
    });
  };

  // BELOW THIS LINE: timer-based updates only

  //var update_timer = $.timer(function() {
    update("/top_games/human_topN", human_leaderboard);
    update("/top_games/ai_topN", ai_leaderboard);
  //}, 1000, true);
});
// vim: et ts=2 sw=2 ai smarttab
