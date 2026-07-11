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

    go-nix-helpers = {
      url = "github:LarsArtmann/go-nix-helpers";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ self, flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.go-nix-helpers.flakeModules.go-standard ];

      go-standard = {
        pname = "template-arch-lint";
        vendorHash = "sha256-cn20VAdEtKO1c80/b+AP5Jjzq9uv4MgFmaYhiT4NdjI=";
        description = "Architecture linter template for Go";
        enableTempl = true;
        subPackages = [ "cmd" ];

        extraBuildAttrs = {
          postInstall = ''
            mv $out/bin/cmd $out/bin/template-arch-lint
          '';
        };

        shellExtraEnv = {
          GOPRIVATE = "github.com/LarsArtmann/*,github.com/larsartmann/*";
        };

        devShellExtraPackages = pkgs: [
          pkgs.delve
          pkgs.gotools
          pkgs.gofumpt
        ];
      };
    };
}
