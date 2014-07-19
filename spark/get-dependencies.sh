#!/bin/bash
#
# Copyright 2014 Albert P Tobey <atobey@datastax.com> @AlTobey
#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# This script will download the Ivy jar to CWD, create a 'deps' directory,
# then populate it with the minimal dependencies of the
# spark-cassandra-connector using ivy.
#
# Using ivy to directly download the connector pulls in many more dependencies
# than are needed (including all of the Cassandra server).
#
# Usage: bash ~/src/f7u12/get-dependencies.sh

if [ ! -x bin/spark-submit ] ; then
  echo "(temporary) This must be run from the top of a Spark tarball distribution."
  exit 1
fi

die () { echo "$*"; exit 1; }

ivyver="2.3.0"
ivyjar="ivy-${ivyver}.jar"

scalaver="2.10"
connectorver="1.0.0-beta2"
connectorjar="spark-cassandra-connector_${scalaver}-${connectorver}.jar"

if [ ! -e $ivyjar ] ; then
  curl -o $ivyjar "http://search.maven.org/remotecontent?filepath=org/apache/ivy/ivy/${ivyver}/$ivyjar"
  [ $? -eq 0 ] || die "downloading $ivyjar failed"
fi

ivypath="$(pwd)/$ivyjar"
ivy () { java -jar $ivypath -dependency $* -retrieve "[artifact]-[revision](-[classifier]).[ext]"; }

mkdir -p deps

if [ ! -e "deps/$connectorjar" ] ; then
  curl -o deps/$connectorjar \
    "http://search.maven.org/remotecontent?filepath=com/datastax/spark/spark-cassandra-connector_${scalaver}/${connectorver}/${connectorjar}"
  [ $? -eq 0 ] || die "downloading $connectorjar failed"
fi

cd deps || die "fatal: could not 'cd deps'"

# 2.1.0-beta1 is not currently compatible with the connector
# https://github.com/datastax/spark-cassandra-connector/issues/85
#ivy com.datastax.cassandra cassandra-driver-core 2.1.0-beta1
ivy com.datastax.cassandra cassandra-driver-core 2.0.3 || die "downloading cassandra-driver-core failed"
ivy org.apache.cassandra cassandra-thrift 2.0.9        || die "downloading cassandra-thrift failed"
ivy joda-time joda-time 2.3                            || die "downloading joda-time failed"
ivy org.joda joda-convert 1.6                          || die "downloading joda-convert failed"

# TODO: figure out how to get ivy to skip these in the first place
rm -f *-{sources,javadoc}.jar

# vim: et ts=2 sw=2 ai smarttab
