# set ALL targets to be PHONY
.PHONY: $(shell sed -n -e '/^$$/ { n ; /^[^ .\#][^ ]*:/ { s/:.*$$// ; p ; } ; }' $(MAKEFILE_LIST))

# text output decoration
B="$$(tput bold)"
UN="$$(tput smul)"
NU="$$(tput rmul)"
DIM="$$(tput dim)"

# text output color
RED="$$(tput setaf 1)"
GRN="$$(tput setaf 2)"
YLW="$$(tput setaf 3)"
BLU="$$(tput setaf 4)"
MGT="$$(tput setaf 5)"
CYN="$$(tput setaf 6)"

# reset terminal output
NRM="$$(tput sgr0)"

# Go compiler
GO := go

# LDFLAGS for compressing the binary or setting version information
LDFLAGS := -X main.kionCliVersion=$$(cat VERSION.md) -s -w

default:
	@printf "\n\
	\
	  $(DIM)usage:$(NRM)  $(B)make <command>$(NRM)\n\n\
	\
	  $(DIM)commands:$(NRM)\n\n\
	\
	    $(B)$(BLU)$(UN)Setup:$(NRM)\n\n\
	\
	    $(B)$(GRN)init$(NRM)                 $(GRN)Setup the repository for development$(NRM)\n\n\
	\
	    $(B)$(BLU)$(UN)Development:$(NRM)\n\n\
	\
	    $(B)$(GRN)build$(NRM)                $(GRN)Build the kion binary$(NRM)\n\
	    $(B)$(GRN)build-windows$(NRM)        $(GRN)Build the kion binary for Windows$(NRM)\n\
	    $(B)$(GRN)gofmt$(NRM)                $(GRN)Run gofmt against the repo$(NRM)\n\
	    $(B)$(GRN)lint$(NRM)                 $(GRN)Run golangci-lint against the repo$(NRM)\n\
	    $(B)$(GRN)test$(NRM)                 $(GRN)Run all go tests$(NRM)\n\
	    $(B)$(GRN)coverage$(NRM)             $(GRN)Run all go tests and calculate coverage$(NRM)\n\
	    $(B)$(GRN)gif$(NRM)                  $(GRN)Build the usage gif$(NRM)\n\n\
	\
	    $(B)$(BLU)$(UN)Helpers:$(NRM)\n\n\
	\
	    $(B)$(GRN)install$(NRM)              $(GRN)Build and install the kion binary to /usr/local/bin$(NRM)\n\
	    $(B)$(YLW)install-symlink$(NRM)      $(GRN)Build and $(B)symlink$(NRM)$(GRN) the kion binary to /usr/local/bin$(NRM)\n\
	    $(B)$(RED)clean$(NRM)                $(RED)Delete generated assets and helpers$(NRM)\n\n"

init:
	@printf "${B}${UN}${BLU}Initializing the repo:${NRM}\n"
	cp tools/pre-commit .git/hooks/pre-commit
	chmod 755 .git/hooks/pre-commit

build:
	@printf "${B}${UN}${BLU}Building Kion CLI:${NRM}\n"
	$(GO) build -ldflags "$(LDFLAGS)" -o kion

build-windows:
	@printf "${B}${UN}${BLU}Building Kion CLI for Windows:${NRM}\n"
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o kion_windows.exe

gofmt:
	@printf "${B}${UN}${BLU}Running gofmt:${NRM}\n"
	gofmt -s -w .

lint:
	@printf "${B}${UN}${BLU}Running golangci-lint:${NRM}\n"
	./tools/lint.sh

test:
	@printf "${B}${UN}${BLU}Running go test:${NRM}\n"
	go test -v -coverpkg=./... -coverprofile=profile.cov ./...

coverage: test
	go tool cover -func profile.cov

gif:
	@printf "${B}${UN}${BLU}Building readme gif:${NRM}\n"
	cd doc && asciicast2gif -s 1 -t monokai -w 89 -h 30 kion-cli-usage.cast kion-cli-usage.gif

install: build
	@printf "${B}${UN}${BLU}Installing Kion CLI:${NRM}\n"
	sudo \cp $$(pwd)/kion /usr/local/bin/kion

install-symlink: build
	@printf "${B}${UN}${BLU}Installing Kion CLI:${NRM}\n"
	sudo ln -sf $$(pwd)/kion /usr/local/bin/kion

clean:
	@printf "${B}${UN}${BLU}Cleaning generated assets and helpers:${NRM}\n"
	rm -f kion
	rm -f kion_windows.exe
	rm -f profile.cov
	rm -f tools/golangci-lint
