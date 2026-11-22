.PHONY: test vet list tone capture-play sink-play

test:
	go test -race ./...

vet:
	go vet ./...

list:
	go run ./cmd/spike list

# usage: make tone A=0 B=1
tone:
	go run ./cmd/spike tone $(A) $(B)

# usage: make capture-play C=1 P=1
capture-play:
	go run ./cmd/spike capture-play $(C) $(P)

# usage: make sink-play DEVS="1 2"
sink-play:
	go run ./cmd/spike sink-play $(DEVS)
