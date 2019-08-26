default:
	go test -v

test:
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm -rf coverage.out

clean:
	rm -rf UwdBot