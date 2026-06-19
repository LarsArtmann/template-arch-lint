{
  description = "Architecture linter template for Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    systems.url = "github:nix-systems/default";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      systems,
      treefmt-nix,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import systems;

      imports = [
        treefmt-nix.flakeModule
      ];

      perSystem =
        {
          config,
          pkgs,
          ...
        }:
        {
          treefmt = {
            projectRootFile = "go.mod";
            programs = {
              gofumpt.enable = true;
              goimports.enable = true;
              nixfmt.enable = true;
              templ.enable = true;
            };
          };

          checks.format = config.treefmt.build.check self;

          packages.default = pkgs.buildGoModule {
            pname = "template-arch-lint";
            version = builtins.substring 0 7 (self.shortRev or "dirty");
            src = ./.;
            subPackages = [ "cmd" ];
            vendorHash = "sha256-MHAHkxjYaoPz4KCxnpSg98s11oXele7kLkGvvzDXaBA=";
            postInstall = ''
              mv $out/bin/cmd $out/bin/template-arch-lint
            '';
          };

          devShells = {
            default = pkgs.mkShell {
              name = "template-arch-lint-dev";

              packages = [
                pkgs.go_1_26
                pkgs.golangci-lint
                pkgs.gopls
                pkgs.delve
                pkgs.gotools
                pkgs.gofumpt
                pkgs.templ
              ];

              GOWORK = "off";
              GOPRIVATE = "github.com/LarsArtmann/*,github.com/larsartmann/*";
            };

            ci = pkgs.mkShellNoCC {
              packages = [
                pkgs.go_1_26
                pkgs.golangci-lint
                pkgs.templ
              ];

              GOWORK = "off";
              GOPRIVATE = "github.com/LarsArtmann/*,github.com/larsartmann/*";
            };
          };
        };
    };
}
