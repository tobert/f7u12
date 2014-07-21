#!/bin/bash

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
# Run spark-shell with the cassandra connector jars in its classpath.

if [ ! -x ./bin/spark-shell ] ; then
	echo "This must be run from the top of a Spark tarball distribution."
	exit 1
fi

# fragile: will break when the dependency gets updated
if [ ! -e "./deps/cassandra-driver-core-2.0.3.jar" ] ; then
	echo "Dependencies seem to be missing. Use the f7u12/spark/get-dependencies.sh script to download them."
	exit 1
fi

CP=$(echo $(pwd)/deps/*.jar |sed 's/ /:/g')

exec ./bin/spark-shell --driver-class-path $CP
