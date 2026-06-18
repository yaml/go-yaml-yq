R := https://github.com/makeplus/makes
ROOT := ../..
M ?= $(ROOT)/.cache/makes
$(shell [ -d '$(M)' ] || git clone -q $(R) '$(M)')

include $(M)/init.mk

GO-VERSION := 1.23.12
include $(M)/go.mk
include $(M)/clean.mk
include $(M)/shell.mk

BIN := update
FILE ?= sample.yaml
EXPR ?= .replicas = 3

MAKES-CLEAN := $(BIN)


build: $(GO)
	go build -o $(BIN) .

run:
	go run main.go $(FILE) '$(EXPR)'
