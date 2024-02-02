# set ALL targets to be PHONY
.PHONY: $(shell sed -n -e '/^$$/ { n ; /^[^ .\#][^ ]*:/ { s/:.*$$// ; p ; } ; }' $(MAKEFILE_LIST))

build:
	go build -o kion

gofmt:
	gofmt -s -w .

lint:
	./tools/lint.sh

test:
	go test -v -coverpkg=./... -coverprofile=profile.cov ./...
	go tool cover -func profile.cov

gif:
	cd doc && asciicast2gif -s 1 -t monokai -w 89 -h 29 kion-cli-usage.cast kion-cli-usage.gif

install: build
	sudo ln -sf $$(pwd)/kion /usr/local/bin/kion

clean:
	rm -f kion
	rm -f profile.cov
