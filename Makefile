dev/test:
	go test
run_app:
	go run main.go &
docker/build:
	docker build -t walkthedog_image .
docker/create_container:
	docker run -d --name walkthedog_bot walkthedog_image
docker/remove_container:
	-docker container rm walkthedog_bot
docker/container/start:
	docker container start walkthedog_bot
docker/container/stop:
	docker container stop walkthedog_bot
docker/container/restart:
	docker container restart walkthedog_bot
docker/clean: docker/container/stop docker/remove_container
docker/logs:
	docker logs walkthedog_bot
