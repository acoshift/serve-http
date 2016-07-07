# serve-http

Very small http service for serve static contents.

## Usage

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
