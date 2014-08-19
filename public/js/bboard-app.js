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

// one-time setup, creating elements, etc.
$(function() {
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
         //<span>Average: </span><span id="game1-avg-score">0</span>

    d3.select(game.target + "-score").text(game.score);
    d3.select(game.target + "-turn-id").text(game.sequence);

    game.last_turn = performance.now();
  };

  var game1 = new F7U12(4);
  game1.name = "00:24:44:dc:0b:25";
  game1.uuid = UUIDjs.create(1).toString();
  game1.init(2);
  game1.target = "#game1";
  game1.render(game1.target);

  var game2 = new F7U12(4);
  game2.name = "00:24:44:dc:0b:25"; // TODO: change to other balance board
  game2.uuid = UUIDjs.create(1).toString();
  game2.init(2);
  game2.target = "#game2";
  game2.render(game2.target);

  game3 = new F7U12(4);
  game3.name = "AI";
  game3.uuid = UUIDjs.create(1).toString();
  game3.init(2);
  game3.target = "#game3";
  game3.render(game3.target);

  var stream_bboard = function (game) {
    console.log(game);
    var sock = new WebSocket("ws://" + window.location.host + "/balance_board/" + game.name);
    sock.onerror = function (e) { console.log("socket error", e); };
    sock.onopen = function (e) {
      sock.onmessage = function(msg) {
        var bb = JSON.parse(msg.data);
        console.log(bb);
        move(game, bb.direction);
        console.log(game.serialize());
      };
    };
  };

  // take input from balance board via a somewhat convoluted path
  stream_bboard(game1);
  stream_bboard(game2);

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
