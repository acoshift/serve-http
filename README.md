# serve-http

Very small http service for serve static contents.

## Usage

```bash
$ serve-http --help
NAME:
   serve-http - a very small http service for serve static contents

USAGE:
   serve-http [--fallback] [--index=index.html] [--dir=www] [--port=80] [--script] [--key=APIKEY]

VERSION:
   1.0.0

AUTHOR(S):
   Thanatat Tamtan <acoshift@gmail.com>

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --fallback     enable to serve index file if current file not found
   --index value  (default: "index.html")
   --dir value    (default: "www")
   --port value   (default: 80)
   --script       enable to run .sh as script
   --key value    authenticate key ?key=
   --help, -h     show help
   --version, -v  print the version
```

- Build

```Dockerfile
FROM docker.io/acoshift/serve-http
COPY public /www
```

```sh
docker build -t my-project .
```

- Run

```sh
docker run -p 8080:80 -d my-project
```

or add config

```sh
docker run -p 8080:80 -d my-project --fallback --index=index.html --dir=www
```
