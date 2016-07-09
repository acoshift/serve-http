default:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/serve-http -a -ldflags '-s' main.go
	docker build -t docker.io/acoshift/serve-http .

push:
	docker push docker.io/acoshift/serve-http
