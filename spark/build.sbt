import AssemblyKeys._

name := "f7u12"

scalaVersion := "2.10.4"

organization := "com.datastax"

version := "1.0"

libraryDependencies += "org.apache.spark" % "spark-core_2.10" % "1.0.2" % "provided"

//libraryDependencies += "com.datastax.spark" %% "spark-cassandra-connector" % "1.0.0-rc4" withSources() withJavadoc()
libraryDependencies += "com.datastax.spark" %% "spark-cassandra-connector" % "1.0.0-beta2"

assemblySettings

mergeStrategy in assembly <<= (mergeStrategy in assembly) { (old) =>
    {
        case x if x.startsWith("META-INF/ECLIPSEF.RSA") => MergeStrategy.last
        case x if x.startsWith("META-INF/mailcap") => MergeStrategy.last
        case x if x.startsWith("plugin.properties") => MergeStrategy.last
        case PathList("com", "esotericsoftware", "minlog", xs @ _*) => MergeStrategy.first // kryo vs minlog
        case x => old(x)
    }
}
