
run:
	docker-compose up -d

# run integration tests
tests:
	go test integration_tests/*.go