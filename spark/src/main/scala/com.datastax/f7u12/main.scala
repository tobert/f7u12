package com.datastax.f7u12

/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Note: being my first serious attempt at both Scala and Spark,
 * this is probably not idiomatic Spark or Scala code
 */

import org.apache.spark.{SparkConf, SparkContext}
import org.apache.spark.SparkContext._ // implicit conversions
import org.apache.spark.storage.StorageLevel._
import com.datastax.spark.connector._
import java.lang.Math
import java.util

object F7U12 {
  type UUID = java.util.UUID

  // a single turn of 2048
  case class Grid (
    GameId:    UUID,
    TurnId:    Int,
    OffsetMs:  Double,
    TurnMs:    Double,
    Player:    Option[String],
    Score:     Option[Float],
    TileVal:   Int,
    TileIdx:   Int,
    Direction: String,
    State:     List[Int]
  )

  // define an ordering to sort by score
  val order_by_score = Ordering.by[(UUID, Float), Float](_._2)

  // TODO: refactor duplicated code into functions
  def main(args: Array[String]) {
    val keyspace = "f7u12"
    val server = "localhost"
    val ntop = 10

    val conf = new SparkConf()
    conf.set("spark.cassandra.connection.host", server)
    conf.set("spark.cassandra.connection.keep_alive_ms", "60000")
    val sc = new SparkContext("local[4]", "F7U12", conf)

    val grids = sc.cassandraTable[Grid](keyspace, "grids").cache

    // xform grids into a kv rdd and ask for it to be cached
    val grids_kv = grids.map(g => (g.GameId, g)).cache

    // get the final grid of each game by finding the max turn id
    val games = grids_kv.reduceByKey((a,b) => (if (a.TurnId > b.TurnId) a else b))

    // split out AI games, then find the top 10 scores
    val ai_games  = games.filter(g => (g._2.Player.getOrElse[String]("Unknown") == "AI"))
    val ai_grids = grids.filter(g => (g.Player.getOrElse[String]("Unknown") == "AI"))

    // get the indexes of ai_topN, add the dimension name string, then save to Cassandra
    val ai_scores = ai_games.map(game => (game._1, (game._2.Score.getOrElse[Float](0))))
    val ai_topN = ai_scores.takeOrdered(10)(order_by_score.reverse)
    val ai_topN_with_rank = sc.parallelize(ai_topN).zipWithIndex.map(t => ("ai_topN", t._2 + 1, t._1._1, t._1._2))
        ai_topN_with_rank.saveToCassandra(keyspace, "top_games", Seq("dimension", "rank", "game_id", "score"))

    // split out non-AI games, find some other things, including top 10
    val human_games = games.filter(g => (g._2.Player.getOrElse[String]("Unknown") != "AI"))
    val human_grids = grids.filter(g => (g.Player.getOrElse[String]("Unknown") != "AI"))

    val human_scores = human_games.map(game => (game._1, (game._2.Score.getOrElse[Float](0))))

    // get the indexes of human_topN, add the dimension name string, then save to Cassandra
    val human_topN = human_scores.takeOrdered(10)(order_by_score.reverse)
    val human_topN_with_rank = sc.parallelize(human_topN).zipWithIndex.map(t => ("human_topN", t._2 + 1, t._1._1, t._1._2))
        human_topN_with_rank.saveToCassandra(keyspace, "top_games", Seq("dimension", "rank", "game_id", "score"))

	// get average score per turn across all human games
    val turn_scores = human_grids.map(grid => (grid.TurnId, grid.Score.getOrElse[Float](0)))
    val human_score_avg_by_turn = turn_scores.groupByKey().map(turn => ("human", turn._1, turn._2.reduce(_+_)/turn._2.count(v => true)))
    human_score_avg_by_turn.saveToCassandra("f7u12", "avg_score_by_turn", Seq("name", "turn_id", "avg_score"))

    // get per-game move counts
    val game_dirs = grids_kv.filter(g => g._2.Direction != "").groupByKey().map(g => Seq(g._1, g._2.groupBy(_.Direction).map(g => (g._1, g._2.count(_ => true)))))
        // one more map to transform Seq() to tuple
        game_dirs.map(gd => (gd(0),gd(1))).saveToCassandra("f7u12", "dir_counts", Seq("game_id", "counts"))

    // count the directions globally
    val all_dir_counts = grids.filter(g => g.Direction != "").map(g => (g.Direction, 1)).reduceByKey(_+_)
    val ai_dir_counts = ai_grids.filter(g => g.Direction != "").map(g => (g.Direction, 1)).reduceByKey(_+_)
    val human_dir_counts = human_grids.filter(g => g.Direction != "").map(g => (g.Direction, 1)).reduceByKey(_+_)

    // compute & overwrite all the simple counts to Cassandra in a key[String]/value[Int] table
    sc.parallelize(Seq(
      ("games",       games.count()),
      ("ai_games",    ai_games.count()),
      ("human_games", human_games.count()),
      ("ai_moves",    ai_games.map(g => (g._2.TurnId)).reduce(_ + _)),
      ("human_moves", human_games.map(g => (g._2.TurnId)).reduce(_ + _)),
      ("all_left",    all_dir_counts.lookup("left")(0)),
      ("all_right",   all_dir_counts.lookup("right")(0)),
      ("all_up",      all_dir_counts.lookup("up")(0)),
      ("all_down",    all_dir_counts.lookup("down")(0)),
      ("ai_left",     ai_dir_counts.lookup("left")(0)),
      ("ai_right",    ai_dir_counts.lookup("right")(0)),
      ("ai_up",       ai_dir_counts.lookup("up")(0)),
      ("ai_down",     ai_dir_counts.lookup("down")(0)),
      ("human_left",  human_dir_counts.lookup("left")(0)),
      ("human_right", human_dir_counts.lookup("right")(0)),
      ("human_up",    human_dir_counts.lookup("up")(0)),
      ("human_down",  human_dir_counts.lookup("down")(0))
    )).saveToCassandra(keyspace, "counts", Seq("name", "value"))
  }
}

// vim: et ts=2 sw=2 ai smarttab
