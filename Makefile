build:
	go build -o main ./main.go

fmt:
	go fmt

dep:
	go mod tidy

docker:
	docker build -t gabismartcare/sql-migration .