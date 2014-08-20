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
  // must be lower case
  var bboards = ["00:24:44:dc:0b:25", "00:24:44:ee:56:a2"];
  var width_px = 445; // width of the grid in pixels (TODO: query the DOM instead)

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
      d3.json("/recent", function (data) { DATA.recent_games = data; });
      d3.json("/avg_score_by_turn", function (data) { DATA.avg_score_by_turn = data; });
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

  var get_avg_score = function (turn) {
    return DATA.avg_score_by_turn.filter(function (v) {
      if (v.name === "human" && v.turn_id === turn) {
        return true;
      }
      return false;
    })[0].avg_score;
  };

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

    d3.select(game.target + "-score").text(game.score);
    d3.select(game.target + "-turn-id").text(game.sequence);
    d3.select(game.target + "-avg-score").text(get_avg_score(game.sequence));

    game.last_turn = performance.now();
  };

  // color the whole game div background by pressure with a gradient
  var colorize = function (game, bb) {
    // draw a radial gradient as the background of the f7u12 game div
    // that follows where the center of pressure is on the balance board
    var gx = width_px - Math.floor(width_px * ((bb.lf_pcnt + bb.lr_pcnt)/100));
    var gy = width_px - Math.floor(width_px * ((bb.lf_pcnt + bb.rf_pcnt)/100));
    d3.select(game.target + " .f7u12-grid")
      .style("background", "radial-gradient(ellipse farthest-corner at " + gx + "px " + gy + "px , #ff0000 1%, #fafa00 50%, #4efa00 99%)");
  };

  var game1 = new F7U12(4);
  game1.name = bboards[0];
  game1.uuid = UUIDjs.create(1).toString();
  game1.init(2);
  game1.target = "#game1";
  game1.render(game1.target);

  var game2 = new F7U12(4);
  game2.name = bboards[1];
  game2.uuid = UUIDjs.create(1).toString();
  game2.init(2);
  game2.target = "#game3";
  game2.render(game2.target);

  game3 = new F7U12(4);
  game3.name = "AI";
  game3.uuid = UUIDjs.create(1).toString();
  game3.init(2);
  game3.target = "#game2";
  game3.render(game3.target);
  game3.dash_update = function () {
    try {
      var id = DATA.recent_games.filter(function (d) { return d.player == "AI"; })[0];
      d3.json("/game/" + (id.game_id || "id-missing"), function (data) {
        game3.cells = data[data.length - 1].state;
        game3.update();
      });
    }
    catch (e) { console.log("Exception: ", e); }
  };
  WIDGETS.game3 = game3;

  var update_pressure_line = (function () {
    var now = Date.now();
    var target = "#pressure-svg";
    var entries = 100;
    var initdata = d3.range(entries).map(function () { return 0; });
    var margin = {top: 20, right: 80, bottom: 30, left: 50};
    var width = 600 - margin.left - margin.right;
    var height = 300 - margin.top - margin.bottom;

    var x = d3.scale.linear().domain([0,entries]).range([0, width]); // only use index for x
    var y = d3.scale.linear().domain([0,100]).range([height, 0]); // percentage
    var xaxis = d3.svg.axis().scale(x).orient("bottom");
    var yaxis = d3.svg.axis().scale(y).orient("left");

    var tline = d3.svg.line().interpolate("basis")
      .x(function (d,i) { return x(i); })
      .y(function (d,i) { return y(d); });

    var bline = d3.svg.line()
      .x(function (d,i) { return x(i); })
      .y(function (d,i) { return y(100 - d); })
      .interpolate("basis");

    var svg = d3.select(target).append("svg")
      .attr("width", width + margin.left + margin.right)
      .attr("height", height + margin.top + margin.bottom)
    .append("g")
      .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

    svg.append("g").attr("class", "x axis").attr("transform", "translate(0," + height + ")").call(xaxis);
    svg.append("g").attr("class", "y axis").call(yaxis);

    var colors = d3.scale.category10();

    // might be able to build this with d3 selections too ... another day
    var lines = {};
    lines[bboards[0]] = {
      dfun: bline,
      lf_pcnt: svg.append("g").append("path").datum(initdata).attr("class", "line").attr("d", bline).attr("stroke", colors(0)),
      rf_pcnt: svg.append("g").append("path").datum(initdata).attr("class", "line").attr("d", bline).attr("stroke", colors(1)),
      lr_pcnt: svg.append("g").append("path").datum(initdata).attr("class", "line").attr("d", bline).attr("stroke", colors(2)),
      rr_pcnt: svg.append("g").append("path").datum(initdata).attr("class", "line").attr("d", bline).attr("stroke", colors(3))
    };
    lines[bboards[1]] = {
      dfun: tline,
      lf_pcnt: svg.append("g").append("path").datum(initdata).attr("class", "line").attr("d", tline).attr("stroke", colors(4)),
      rf_pcnt: svg.append("g").append("path").datum(initdata).attr("class", "line").attr("d", tline).attr("stroke", colors(5)),
      lr_pcnt: svg.append("g").append("path").datum(initdata).attr("class", "line").attr("d", tline).attr("stroke", colors(6)),
      rr_pcnt: svg.append("g").append("path").datum(initdata).attr("class", "line").attr("d", tline).attr("stroke", colors(7))
    };

    // return a function that can be called on websocket events to update the lines
    // by mac & sensor
    return function (bb) {
      if (!lines.hasOwnProperty(bb.mac_address)) {
        console.log("Unrecognized MAC address: ", bb.mac_address, bb);
        return;
      }

      ["lf_pcnt", "rf_pcnt", "lr_pcnt", "rr_pcnt"].forEach(function (key) {
        var board = lines[bb.mac_address];
        var line = board[key];
        var data = line.datum();
            data.shift();
            data.push(bb[key]);
        line.datum(data).attr("d", board.dfun);
      });
    };
  }());

  var stream_bboard = function (game) {
    var sock = new WebSocket("ws://" + window.location.host + "/balance_board/" + game.name);
    sock.onerror = function (e) { console.log("socket error", e); };
    sock.onopen = function (e) {
      sock.onmessage = function(msg) {
        var bb = JSON.parse(msg.data);
        bb.timestamp = Date.parse(bb.timestamp);

        // the backend can send zeroed messages if no recent rows are in Cassandra
        // in that case, the board is probably idle, so assume even distribution
        if (bb.direction === "") {
          bb.lf_pcnt = 25;
          bb.rf_pcnt = 25;
          bb.lr_pcnt = 25;
          bb.rr_pcnt = 25;
        }

        // direction may be 'none', unsupported by f7u12
        if (bb.direction == "up" || bb.direction == "down" || bb.direction == "left" || bb.direction == "right") {
          move(game, bb.direction);
        }

        colorize(game, bb);

        update_pressure_line(bb);
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
