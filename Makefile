default:
	go build ./cmd/uwdbot

test:
	go test -v ../...

coverage:
	go test ../... -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm -rf coverage.out

hooks:
	git config core.hooksPath hooks

clean:
	rm -rf ./cmd/uwdbot/uwdbot