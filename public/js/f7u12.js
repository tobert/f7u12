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
 * F7U12: A simple implementation of the 2048 game using D3.
 */

// constructor: takes a width and identifier for the target div
var F7U12 = function (width, target) {
  this.width = width;
  this.size = width * width;
  this.last = this.size - 1;
  this.sequence = 0;
  this.cells = new Array(this.size);

  for (var i=0; i<this.size; i++) {
    this.cells[i] = 0;
  }

  this.target = target;
  this.container = d3.select(target);
  this.container.classed("f7u12-grid", true);
};

F7U12.cell_class = function (d,i) {
  if (d === 0) {
    return "f7u12-cell f7u12-cell-empty";
  } else {
    return "f7u12-cell f7u12-cell-" + d;
  }
};

// used internally in a few .text() calls below
F7U12.print = function (d) {
  if (d > 0) {
    return d;
  } else {
    return "";
  }
};

// makes a copy of the object
F7U12.prototype.clone = function () {
  var out = new F7U12(this.width, this.target);
  this.cells.forEach(function (d,i) {
    out.cells[i] = d;
  });
};

// render the game grid
F7U12.prototype.render = function () {
  var game = this;
  var cells = game.container.selectAll(".f7u12-cell")
    .data(game.cells)
    .enter()
    .append("div")
      .attr("class", F7U12.cell_class)
      .attr("data-id", function (d,i) { return i; })
      .attr("style", function (d,i) {
         if (i % game.width == 0) { return "clear: both;" }
      })
    .text(F7U12.print);

  // set up transitions on value changes, doesn't do much yet
  cells.transition().tween("text", function(d,i) {
    var old = parseInt(this.textContent) || 0;

    if (d === 0 && old === 0) {
      return null;
    }

    var i = d3.interpolateRound(old, d);
    return function(t) {
      this.textContent = i(t);
    };
  });
};

// calculate the from/to coordinates for use in slide() & merge()
// eol is set to true on pairs that wrap around the grid
// 'this' is expected to be the F7U12 object
F7U12.next = function (dir, i) {
    var out = { from: 0, to: 0, eol: false };

    // DOWN:
    // the cell below i is width cells forward
    if (dir === "down") {
      out.from = i;
      out.to = i + this.width;
      if (out.to >= this.size) {
        out.eol = true;
      }
    }

    // UP:
    // start on the bottom right
    // the cell above from is width cells backward
    else if (dir === "up") {
      out.from = this.last - i;
      out.to = out.from - this.width;
      if (out.to < 0) {
        out.eol = true;
      }
    }

    // RIGHT:
    // the cell to the right is forward one cell
    else if (dir === "right") {
      out.from = i;
      out.to = i + 1;

      if (out.to % this.width === 0) {
        out.eol = true;
      }
    }

    // LEFT:
    // start at the bottom right
    // the cell to the left is backwards one cell
    else if (dir === "left") {
      out.from = this.last - i;
      out.to = out.from - 1;

      if (out.from % this.width === 0) {
        out.eol = true;
      }
    }

    // should never happen normally, most likely a typo in input wiring
    else {
      console.log("BUG: invalid direction '" + dir + "'.");
    }

    return out;
};

// move populated tiles into unpopulated tiles in the given direction
// calls itself recursively until there are no more moves
// return the number of slides completed
F7U12.prototype.slide = function (dir, prev_count) {
  var count = prev_count || 0;
  var game = this;
  var dirty = false;

  var updated = game.cells.map(function (val, i) {
    var idxs = F7U12.next.call(game, dir, i);

    if (idxs.eol) {
      return val;
    }

    if (game.cells[idxs.to] == 0 && game.cells[idxs.from] > 0) {
      game.cells[idxs.to] = game.cells[idxs.from];
      game.cells[idxs.from] = 0;
      count++;
      dirty = true;
    }

    return game.cells[i];
  });

  game.container.selectAll(".f7u12-cell")
    .data(updated)
    .attr("class", F7U12.cell_class)
    .text(F7U12.print);

  if (dirty) {
    game.slide(dir, count);
  }

  return count;
};

// merge cells with matching numbers
// returns a list of indices of cells merged
F7U12.prototype.merge = function (dir) {
  var game = this;

  // track which cells have been updated and only merge them once per move
  var merged = new Array(game.size);
  merged.forEach(function (v,i) { merged[i] = 0; });

  var updated = game.cells.map(function (val, i) {
    var idxs = F7U12.next.call(game, dir, i);

    if (idxs.eol) {
      return val;
    }

    if (game.cells[idxs.from] == game.cells[idxs.to] && !merged[idxs.from]) {
      game.cells[idxs.from] = 0;
      game.cells[idxs.to] = game.cells[idxs.to] * 2;
      merged[idxs.to] = game.cells[idxs.to]; // return the combined value
    }

    return game.cells[i];
  });

  game.container.selectAll(".f7u12-cell")
    .data(updated)
    .attr("class", F7U12.cell_class)
    .text(F7U12.print);

  return merged.filter(function (val) { if (val > 0) { return true; } else { return false; } });
};

// collapse tiles, move, collapse again
// returns true if the board changed, false otherwise
// does NOT insert a new piece, you must call insert()
F7U12.prototype.move = function (dir) {
  var game = this;

  var slides = game.slide(dir);
  var merged = game.merge(dir);
  slides += game.slide(dir);

  if (slides > 0 || merged.length > 0) {
    game.sequence++;
    return true;
  }

  return false;
};

// randomly insert a 2 or 4 on the grid
F7U12.prototype.insert = function () {
  var game = this;

  // make a list of all empty cells (value = 0)
  var available = game.cells.map(function (val, i) {
    if (val == 0) {
      return i;
    }
  }).filter(function (val) { return val != undefined; });

  // shuffle the empty index list
  d3.shuffle(available);
  // take the first value and assign 2 or 4 at random
  game.cells[available[0]] = d3.shuffle([2,4])[0];

  game.container.selectAll(".f7u12-cell")
    .data(game.cells)
    .attr("class", F7U12.cell_class)
    .text(F7U12.print);

  // return the index of the inserted tile
  return available[0];
};

// render the board with count values on it
// returns the indexes of the tiles inserted
F7U12.prototype.init = function (count) {
  this.render();
  var tiles = [];
  for (var i=0; i<count; i++) {
    tiles.push(this.insert());
  }
  return tiles;
};
