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

  // there are two timers, one fires XHRs to update data stored in DATA
  // the other updates content on the page using data in DATA
  // These will be out of sync a little but it should be no big deal
  // since the code below is written to keep them independent.
  // This should probably be replaced by some kind of library that
  // hooks up to websockets nicely, but for now this is good enough.
  var DATA = {};
  DATA.update = function () {
      d3.json("/counts", function (data) { DATA.counts = data; split_counts(data); });
      d3.json("/top_games/human_topN", function (data) { DATA.human_topn = data; });
      d3.json("/top_games/ai_topN", function (data) { DATA.ai_topn = data; });
      d3.json("/recent", function (data) { DATA.recent_games = data; });
  };
  DATA.update();

  // place all widgets here that should auto update
  var WIDGETS = {};

  // find all keys with a dash_update() method then call it
  WIDGETS.update = function () {
    d3.keys(WIDGETS).forEach(function (key) {
      if (typeof WIDGETS[key].dash_update === 'function') {
        WIDGETS[key].dash_update(WIDGETS[key]);
      }
    });
  };

  // use jquery.timer.js to update every second
  // TODO: add play/pause buttons?
  // these will run a litle out of sync and that's OK for now
  DATA.timer = $.timer(DATA.update, 1000, true);
  WIDGETS.timer = $.timer(WIDGETS.update, 200, true);

  // returns an updater that can be attached to games
  var make_grid_updater = function (game) {
    return function () {
      if (typeof DATA.recent_games != "object" || typeof game == "undefined") {
        return;
      };

      var id = DATA.recent_games.filter(function (d) {
        // hacks: these should never be undef
        if (typeof d == "undefined") {
          return false;
        }
        if (d.hasOwnProperty("player") && game.hasOwnProperty("player")) {
          return d.player == game.player
        }
        return false;
      })[0];

      // probably the first update
      if (id == undefined) {
        console.log("No game ID fround for player:", game.player);
        return;
      }

      d3.json("/game/" + (id.game_id || "id-missing"), function (data) {
        game.cells = data[data.length - 1].state;
        game.update();
      });
    };
  };

  // TODO: add the human grid and align it
  WIDGETS.grid1 = new F7U12(4);
  WIDGETS.grid1.player = "player1";
  WIDGETS.grid1.render("#grid1");
  WIDGETS.grid1.dash_update = make_grid_updater(WIDGETS.grid1);

  WIDGETS.grid2 = new F7U12(4);
  WIDGETS.grid2.player = "player2";
  WIDGETS.grid2.render("#grid2");
  WIDGETS.grid2.dash_update = make_grid_updater(WIDGETS.grid2);

  WIDGETS.grid3 = new F7U12(4);
  WIDGETS.grid3.player = "AI";
  WIDGETS.grid3.render("#grid3");
  WIDGETS.grid3.dash_update = make_grid_updater(WIDGETS.grid3);

  d3.json("/top_games/human_topN", function (data) {
    DATA.human_topn = data;
    WIDGETS.human_leaderboard = new F7U12.Leaderboard("#human_topN_leaderboard", "Human Top 10");
    WIDGETS.human_leaderboard.render(DATA.human_topn);
    WIDGETS.human_leaderboard.dash_update = function () {
      WIDGETS.human_leaderboard.update(DATA.human_topn);
    };
  });

  d3.json("/top_games/ai_topN", function (data) {
    DATA.ai_topn = data;
    WIDGETS.ai_leaderboard = new F7U12.Leaderboard("#ai_topN_leaderboard", "AI Top 10");
    WIDGETS.ai_leaderboard.render(data);
    WIDGETS.ai_leaderboard.dash_update = function () {
      WIDGETS.ai_leaderboard.update(DATA.ai_topn);
    };
  });

  var svg1 = dimple.newSvg("#graph1", 600, 200);
  var svg2 = dimple.newSvg("#graph2", 600, 200);

  var split_counts = function (data) {
    var re = /^(ai|human)_(up|down|left|right)/;
    var ai_re = /^ai_(up|down|left|right)/;
    var human_re = /^human_(up|down|left|right)/;

    DATA.ai_moves = data.filter(function (d) { return ai_re.test(d.name); })
      .map(function (d) { return { name: d.name.replace(re, "$2"), value: d.value }; });

    DATA.human_moves = data.filter(function (d) { return human_re.test(d.name); })
      .map(function (d) { return { name: d.name.replace(re, "$2"), value: d.value }; });
  };

  d3.json("/counts", function (data) {
    split_counts(data);

  //var svg2 = dimple.newSvg("#graph2", 590, 400);
    WIDGETS.ai_move_chart = new dimple.chart(svg2, DATA.ai_moves);
    WIDGETS.ai_move_chart.setBounds(60, 30, 440, 130);
    WIDGETS.ai_move_chart.addCategoryAxis("x", "name");
    WIDGETS.ai_move_chart.addMeasureAxis("y", "value");
    WIDGETS.ai_move_chart.addSeries("value", dimple.plot.bar);
    WIDGETS.ai_move_chart.draw();
    // TODO: actually wire up data in the update
    WIDGETS.ai_move_chart.dash_update = function () {
      WIDGETS.ai_move_chart.data = DATA.ai_moves;
      WIDGETS.ai_move_chart.draw();
    };

    WIDGETS.human_move_chart = new dimple.chart(svg1, DATA.human_moves);
    WIDGETS.human_move_chart.setBounds(60, 30, 440, 130);
    WIDGETS.human_move_chart.addCategoryAxis("x", "name");
    WIDGETS.human_move_chart.addMeasureAxis("y", "value");
    WIDGETS.human_move_chart.addSeries("value", dimple.plot.bar);
    WIDGETS.human_move_chart.draw();
    // TODO: actually wire up data in the update
    WIDGETS.human_move_chart.dash_update = function () {
      WIDGETS.human_move_chart.data = DATA.human_moves;
      WIDGETS.human_move_chart.draw();
    };

    // update both from here to save a little code
    WIDGETS.bignums = {};
    WIDGETS.bignums.dash_update = function () {
      ["human_games", "human_moves", "ai_games", "ai_moves"].forEach(function (name) {
        d3.select("#" + name)
          .data(DATA.counts.filter(function (d) { return d.name === name; }))
          .text(function (d) { return d.value; });
      });
    };
    WIDGETS.bignums.dash_update();
  });
});
// vim: et ts=2 sw=2 ai smarttab
