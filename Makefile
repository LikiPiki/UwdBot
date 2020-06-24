default:
	go build

test:
	go test -v

coverage:
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm -rf coverage.out

hooks:
	git config core.hooksPath hooks

clean:
	rm -rf UwdBot