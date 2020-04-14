# Book REST API
A [Go](https://golang.org/) example of a RESTful API architecture, using [Chi](https://github.com/go-chi/chi) and [upper-db](http://upper.io/db.v3/).

## Running
Just launch Docker Compose.
```shell
docker-compose up -d
```
If you'd like to see logs (add `-f` to follow the logs, then CTRL+C to quit).
```shell
docker-compose logs
```
Finally, shutdown (cut off `-v` if you want to keep database data).
```shell
docker-compose down -v
```
