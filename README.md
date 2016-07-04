# serve-http

Very small http service for serve static contents.

## Usage

- Build

```Dockerfile
FROM docker.io/acoshift/serve-http
COPY public /public
```

```sh
docker build -t my-project .
```

- Run

```sh
docker run -p 8080:80 -d my-project
```

### Fallback

If you want to fallback `index.html` when file not found, you can use fallback image

```Dockerfile
FROM docker.io/acoshift/serve-http-fallback
```
