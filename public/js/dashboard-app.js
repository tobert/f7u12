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
  document.body.addEventListener('touchmove', function(event) { event.preventDefault(); }, false);

  var human_leaderboard = new F7U12.Leaderboard("#human_topN_leaderboard", "Human Top 10");
  var ai_leaderboard = new F7U12.Leaderboard("#ai_topN_leaderboard", "AI Top 10");

  $.ajax({
    url: "/top_games/human_topN",
    dataType: "json",
    success: function (data) {
      console.log(data);
      human_leaderboard.render(data);
    }
  });

  $.ajax({
    url: "/top_games/ai_topN",
    dataType: "json",
    success: function (data) {
      console.log(data);
      ai_leaderboard.render(data);
    }
  });
});

// vim: et ts=2 sw=2 ai smarttab
