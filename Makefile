.PHONY: test list tone

test:
	go test -race ./internal/engine/...

list:
	go run ./cmd/spike list

# usage: make tone A=0 B=1
tone:
	go run ./cmd/spike tone $(A) $(B)
