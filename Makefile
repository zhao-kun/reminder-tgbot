
SHELL := /bin/bash

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))
SOURCE_FILES := $(shell find ${MKFILE_DIR}{repo,server,client,model} -type f -name "*.go")


tgbot: ${SOURCE_FILES} tgbot.go
	go build -o $@ tgbot.go
