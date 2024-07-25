proto-sso:
	cd ./protos
	protoc -I proto proto/sso/sso.proto --go_out=./gen/go --go_opt=paths=source_relative --go-grpc_out=./gen/go --go-grpc_opt=paths=source_relative
migrate:
	go run ./cmd/migrations
run:
	go run ./cmd/sso --config=./config/local.yaml
test-logger:
	go run ./cmd/test