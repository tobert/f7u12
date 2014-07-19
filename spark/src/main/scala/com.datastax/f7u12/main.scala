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
  // a single turn of 2048
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
  )

  // define an ordering to sort by score
  val order_by_score = Ordering.by[(java.util.UUID, Float), Float](_._2)

  // TODO: refactor duplicated code into functions
  def main(args: Array[String]) {
    val keyspace = "f7u12"
    val server = "zorak"
    val ntop = 10

    val conf = new SparkConf()
    conf.set("cassandra.connection.host", server)
    conf.set("cassandra.connection.keep_alive_ms", "60000")
    val sc = new SparkContext("local[4]", "F7U12", conf)

    val grids = sc.cassandraTable[Grid](keyspace, "grids").cache

    // xform grids into a kv rdd and ask for it to be cached
    val games = grids.map(g => (g.GameId, g)).cache

    // get the final grid of each game by finding the max turn id
    val final_grids = games.reduceByKey((a,b) => (if (a.TurnId > b.TurnId) a else b))

    // split out AI games, then find the top 10 scores
    val ai_games  = final_grids.filter(g => (g._2.Player.getOrElse[String]("Unknown") == "AI"))

    // get the indexes of ai_topN, add the dimension name string, then save to Cassandra
    val ai_scores = ai_games.map(game => (game._1, (game._2.Score.getOrElse[Float](0))))
    val ai_topN = ai_scores.takeOrdered(10)(order_by_score.reverse)
    val ai_topN_with_rank = sc.parallelize(ai_topN).zipWithIndex.map(t => ("ai_topN", t._2 + 1, t._1._1, t._1._2))
        ai_topN_with_rank.saveToCassandra(keyspace, "top_games", Seq("dimension", "rank", "game_id", "score"))

    // split out non-AI games, find some other things, including top 10
    val human_games = final_grids.filter(g => (g._2.Player.getOrElse[String]("Unknown") != "AI"))

    // analyze human move latency
    val human_scores = human_games.map(game => (game._1, (game._2.Score.getOrElse[Float](0))))
    val human_grids = grids.filter(g => (g.Player.getOrElse[String]("Unknown") != "AI")).collect
    val human_move_lat = human_grids.map(g => (g.Player, g.TurnMs))

    // get the indexes of human_topN, add the dimension name string, then save to Cassandra
    val human_topN = human_scores.takeOrdered(10)(order_by_score.reverse)
    val human_topN_with_rank = sc.parallelize(human_topN).zipWithIndex.map(t => ("human_topN", t._2 + 1, t._1._1, t._1._2))
        human_topN_with_rank.saveToCassandra(keyspace, "top_games", Seq("dimension", "rank", "game_id", "score"))

    // compute & overwrite all the simple counts to Cassandra in a key[String]/value[Int] table
    sc.parallelize(Seq(
      ("games",       final_grids.count()),
      ("ai_moves",    ai_games.map(g => (g._2.TurnId)).reduce(_ + _)),
      ("human_moves", human_games.map(g => (g._2.TurnId)).reduce(_ + _))
    )).saveToCassandra(keyspace, "counts", Seq("name", "value"))
  }
}

// vim: et ts=2 sw=2 ai smarttab
