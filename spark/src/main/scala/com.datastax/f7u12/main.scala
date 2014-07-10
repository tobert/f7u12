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
import com.datastax.driver.spark._

object F7u12 {
  def main(args: Array[String]) {
		val conf = new SparkConf()
		conf.set("cassandra.connection.host", "127.0.0.1")
		val sc = new SparkContext("local[4]", "F7U12", conf)

		val grids = sc.cassandraTable("f7u12", "grids")
		grids.count
  }
}
