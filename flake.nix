{
  description = "Architecture linter template for Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    go-nix-helpers = {
      url = "github:LarsArtmann/go-nix-helpers";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.go-nix-helpers.flakeModules.go-standard ];

      go-standard = {
        pname = "template-arch-lint";
        vendorHash = "sha256-nqKrLvui5XlX+n51DQ/w+F+dsTLu3efH7XL6vLv9oyM=";
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
