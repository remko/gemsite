DOMAIN ?= g.mko.re

GO_TOOLS_BIN_DIR=$(shell go env GOPATH)/bin
REFLEX = $(GO_TOOLS_BIN_DIR)/reflex
STATICCHECK=$(GO_TOOLS_BIN_DIR)/staticcheck
VULNCHECK=$(GO_TOOLS_BIN_DIR)/govulncheck

ifeq ($(BUILD_RPI),1)
BUILD_ENV=env GOOS=linux GOARCH=arm GOARM=6
endif

all: build

build: search.idx server.crt server.key
	$(BUILD_ENV) go build ./cmd/servegemsite

.PHONY: search.idx
search.idx: buildgemsite
	./buildgemsite

buildgemsite: $(wildcard cmd/buildgemsite/*.go)
	go build ./cmd/buildgemsite

dev:
	$(REFLEX) -r '(^(templates|gemsite|content)/.*|\.go$$)' -s -- sh -c "make build && ./servegemsite"

install-tools:
	go install github.com/cespare/reflex@latest 
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

.PHONY: lint
lint:
	go vet ./...
	$(STATICCHECK) -checks inherit  ./... 

.PHONY: audit
audit:
	$(VULNCHECK) ./... 

.PHONY: gen-cert
gen-cert server.crt server.key:
	openssl req -x509 -newkey rsa:4096 -sha256 -days 3650 -nodes -keyout server.key -out server.crt -subj "/CN=$(DOMAIN)" -addext "subjectAltName=DNS:$(DOMAIN),DNS:localhost,IP:127.0.0.1"

admin.crt admin.key:
	openssl req       -newkey rsa:4096 -sha256 -days 3650 -nodes -keyout admin.key  -out admin-req.crt  -subj "/CN=admin@$(DOMAIN)"
	openssl x509 -req -in admin-req.crt -days 3650 -CA server.crt -CAkey server.key -set_serial 01 -out admin.crt
	rm -f admin-req.crt
