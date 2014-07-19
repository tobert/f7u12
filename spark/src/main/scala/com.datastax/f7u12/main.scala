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
 */

import org.apache.spark.{SparkConf, SparkContext}
import org.apache.spark.SparkContext._ // implicit conversions
import org.apache.spark.storage.StorageLevel._
import com.datastax.spark.connector._
import java.lang.Math
import java.util

object F7U12 {
  case class Grid (
    GameId:    java.util.UUID,
    TurnId:    Int,
    OffsetMs:  Double,
    TurnMs:    Double,
    Player:    Option[String],
    Score:     Option[Float],
    TileVal:   Int,
    TileIdx:   Int,
    Direction: String,
    State:     List[Int]
  );

  // define an ordering to sort by score
  val order_by_score = Ordering.by[(java.util.UUID, Float), Float](_._2)

  def main(args: Array[String]) {
    val conf = new SparkConf()
    conf.set("cassandra.connection.host", "zorak")
    conf.set("cassandra.connection.keep_alive_ms", "60000")
    val sc = new SparkContext("local[4]", "F7U12", conf)

    val grids = sc.cassandraTable[Grid]("f7u12", "grids") //.cache

    // xform grids into a kv rdd and ask for it to be cached
    val games = grids.map(g => (g.GameId, g))

    // get the final grid of each game by finding the max turn id
    val final_grids = games.reduceByKey((a,b) => (if (a.TurnId > b.TurnId) a else b))

    // split out AI games, then find the top 10 scores
    val ai_games  = final_grids.filter(g => (g._2.Player.getOrElse[String]("Unknown") == "AI"))
    val ai_scores = ai_games.map(game => (game._1, (game._2.Score.getOrElse[Float](0))))
    val ai_top10 = ai_scores.takeOrdered(10)(order_by_score.reverse)
    val ai_moves = ai_games.map(g => (g._2.TurnId)).reduce(_ + _)

    // split out non-AI games, find some other things, including top 10
    val human_games = final_grids.filter(g => (g._2.Player.getOrElse[String]("Unknown") != "AI"))
    val human_scores = human_games.map(game => (game._1, (game._2.Score.getOrElse[Float](0))))
    val human_top10 = human_scores.takeOrdered(10)(order_by_score.reverse)
    val human_moves = human_games.map(g => (g._2.TurnId)).reduce(_ + _)

    // analyze human move latency
    val human_grids = grids.filter(g => (g.Player.getOrElse[String]("Unknown") != "AI")).collect
    val human_move_lat = human_grids.map(g => (g.Player, g.TurnMs))
  }
}

// vim: et ts=2 sw=2 ai smarttab
