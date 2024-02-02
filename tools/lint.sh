#!/bin/bash
#
# Lint
#
# Script for running golangci-lint locally.
#
# Author(s): Cody Buell
#
# Usage: run via the makefile `make lint`

# color helpers
BOLD="$(tput -Txterm-256color bold)"
UNDER="$(tput -Txterm-256color smul)"
GREEN="$(tput -Txterm-256color setaf 2)"
YELLOW="$(tput -Txterm-256color setaf 3)"
BLUE="$(tput -Txterm-256color setaf 4)"
NORM="$(tput -Txterm-256color sgr0)"

# create the tools dir if it does not exist
if [ ! -d tools ]; then
  mkdir tools
fi

# check local go version matches what we are running in the pipe, warn if not
LOCAL_GO_VERSION=$(go version | { read -r _ _ v _; echo "${v#go}"; })
TARGET_GO_VERSION="1.21.6"
if [ "$LOCAL_GO_VERSION" != "$TARGET_GO_VERSION" ]; then
  echo
  echo "${YELLOW}You are running ${BOLD}Go ${LOCAL_GO_VERSION}${NORM}${YELLOW} locally but the pipeline is on ${BOLD}Go ${TARGET_GO_VERSION}${NORM}${YELLOW}.${NORM}"
  echo "${YELLOW}This can cause inconsistent findings between the pipe and this local${NORM}"
  echo "${YELLOW}run. ${UNDER}Please install Go ${TARGET_GO_VERSION} locally.${NORM}"
  echo
fi

# grab the expected golangci-lint version
GOLANG_CI_LINT_VERSION="v1.55.2"
# GOLANG_CI_LINT_VERSION=$(awk '/GOLANG_CI_LINT_VERSION:/{print $2}' < .gitlab-ci.yml)
echo "${YELLOW}Expecting golangci-lint version:  ${BOLD}${BLUE}${GOLANG_CI_LINT_VERSION}${NORM}"

# find install path
if [ -f "tools/golangci-lint" ]; then
  LINTER_PATH="tools/golangci-lint"
elif which golangci-lint > /dev/null; then
  LINTER_PATH=$(which golangci-lint)
else
  LINTER_PATH="none"
fi

# detect version
if [ "${LINTER_PATH}" == "none" ]; then
  LOCAL_GOLANGCI_LINT_VER="not installed"
else
  LOCAL_GOLANGCI_LINT_VER=$($LINTER_PATH --version 2> /dev/null | awk '{print "v"$4}')
fi
echo "${YELLOW}Found golangci-lint version:      ${BOLD}${BLUE}${LOCAL_GOLANGCI_LINT_VER}${NORM}"
echo

if [ "$LOCAL_GOLANGCI_LINT_VER" != "$GOLANG_CI_LINT_VERSION" ]; then
  echo "${YELLOW}Installing the expected version of golangci-lint in $(pwd)/tools${NORM}"
  echo
  wget -O /tmp/golangci-lint_install.sh https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh &> /dev/null
  bash /tmp/golangci-lint_install.sh -b tools "$GOLANG_CI_LINT_VERSION"
  LINTER_PATH="tools/golangci-lint"
  echo
fi

# run the linter
echo "${YELLOW}Linting...${NORM}"
echo
${LINTER_PATH} run --issues-exit-code 1

# check exit code
EXITCODE=$?
if [ "$EXITCODE" != "0" ]; then
  echo
  exit $EXITCODE
else
  echo "${GREEN}No findings!${NORM}"
  echo
fi
