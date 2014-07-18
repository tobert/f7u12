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

// spark/blob/master/sql/catalyst/src/main/scala/org/apache/spark/sql/catalyst/ScalaReflection.scala
// needs support for java.util.UUID or MatchError, using String for now

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

  def main(args: Array[String]) {
    val conf = new SparkConf()
    conf.set("cassandra.connection.host", "127.0.0.1")
    conf.set("cassandra.connection.keep_alive_ms", "60000")
    val sc = new SparkContext("local[4]", "F7U12", conf)

    val grids = sc.cassandraTable[Grid]("f7u12", "grids")

    //val tiles = grids.map(g => (g.GridId, g))
    // (grid_id, max_tile_value)
    //val max_tiles = tiles.map(row => (row._1, row._2.reduce((a,b) => Math.max(a,b))))
    //val count = max_tiles.filter(r => r._2 == 1024).count
  }
}
