name := "f7u12-spark"

scalaVersion := "2.10.4"

organization := "com.datastax"

version := "1.0"

scalacOptions ++= Seq("-encoding", "UTF-8")

libraryDependencies <<= scalaVersion {
  scala_version => Seq(
    "org.apache.spark" %% "spark-core" % "1.0.0",
    "org.apache.spark" %% "spark-streaming" % "1.0.0"
  )
}

libraryDependencies += "com.datastax.cassandra" % "cassandra-driver-core" % "2.0.3"

resolvers += "Tobert Repo" at "http://tobert.org/mvn"

libraryDependencies += "com.datastax.cassandra" %% "cassandra-driver-spark" % "1.0.0-SNAPSHOT"

