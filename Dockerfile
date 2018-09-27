FROM golang:1.11

WORKDIR /code

COPY . .

RUN ["go", "get", "github.com/githubnemo/CompileDaemon"]


ENTRYPOINT CompileDaemon -log-prefix=false -graceful-kill=true -build="go build -o go-movies-json ./" -command="./go-movies-json"
