.PHONY: server
server:
	go run ./cmd/server/main.go

.PHONY: wasm
wasm:
	GOOS=js GOARCH=wasm go build  -o ./assets/main.wasm ./src/core/wasm.go

.PHONY: shell
shell:
	nix develop -c $$SHELL


