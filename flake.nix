{
  inputs = {
    systems.url = "github:nix-systems/default";
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
  };

  outputs = { self, nixpkgs, systems, }:
    let
      forEachSystem = f:
        nixpkgs.lib.genAttrs (import systems) (system:
          f {
            pkgs = import nixpkgs {
              config.allowUnfree = true;
              inherit system;
            };
          });
    in {
      devShells = forEachSystem ({ pkgs }: {
        default = pkgs.mkShellNoCC {
          packages = with pkgs; [
            go_1_24
            gopls
            gotools

            # AI/ML tools
            nodejs-slim

            # Development essentials
            git
            gnumake
            direnv
            fd
            ripgrep
            uutils-coreutils-noprefix

            # Code formatter stuffs
            prettierd
            deno
            vscode-langservers-extracted
          ];
        };
      });
    };
}
