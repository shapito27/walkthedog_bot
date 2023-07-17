test:
	go test
run_bot:
	go run main.go
docker_build_container:
	docker build -t walkthedog_image .
docker_run_container:
	docker run -d --name walkthedog_bot walkthedog_image
docker_logs:
	docker logs walkthedog_bot
