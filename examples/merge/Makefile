R := https://github.com/makeplus/makes
ROOT := ../..
M ?= $(ROOT)/.cache/makes
$(shell [ -d '$(M)' ] || git clone -q $(R) '$(M)')

include $(M)/init.mk

GO-VERSION := 1.23.12
include $(M)/go.mk
include $(M)/clean.mk
include $(M)/shell.mk

BIN := merge
FILE1 ?= base.yaml
FILE2 ?= overlay.yaml

MAKES-CLEAN := $(BIN)


build: $(GO)
	go build -o $(BIN) .

run: $(GO)
	go run main.go $(FILE1) $(FILE2)
