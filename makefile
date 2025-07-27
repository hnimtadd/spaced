.PHONY: server
server:
	nix-shell -p http-server --command 'http-server -c-1 -p 8080'

.PHONY: wasm
wasm:
	GOOS=js GOARCH=wasm go build  -o ./assets/main.wasm ./src/core/wasm.go

.PHONY: shell
shell:
	nix develop -c $$SHELL


