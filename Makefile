export REDIS_HOST = localhost

# Usage: `make test ARGS=./infrastructure/persistance/views -run TestFunc`.
TEST_ARGS := $(if $(ARGS),$(ARGS),./...)

proto:
	protoc --go_out=. --go_opt=paths=source_relative protobuf/event.proto

up:
	${MAKE} down
	docker-compose up -d

bare_test:
	go test $(TEST_ARGS)

down:
	-docker-compose down

test: | up bare_test down
