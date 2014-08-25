f7u12
=====

This is a cut-down version of f7u12 without Cassandra or Spark
integration. It's just the 2048 game by itself. I kept Go webserver
since it's there and lets me keep the same FROM image as the full
app.

Usage
=====

Pull down the Docker image and browse to http://localhost:8080

```
docker pull tobert/f7u12-2048
docker run -d -p 127.0.0.1:8080:8080 tobert/f7u12-2048
```
