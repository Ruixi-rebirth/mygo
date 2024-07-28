{
  description = "A cargo-like tool for Go projects";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "mygo";
          version = "git";

          src = ./.;

          subPackages = [ "." ];

          vendorHash = null;

          buildInputs = [ pkgs.go ];

          postInstall = ''
            mkdir -p $out/share/mygo/completions
            $out/bin/mygo completion bash > $out/share/mygo/completions/mygo-completion.bash
            $out/bin/mygo completion zsh > $out/share/mygo/completions/mygo-completion.zsh
            $out/bin/mygo completion fish > $out/share/mygo/completions/mygo-completion.fish
          '';
        };

        devShell = pkgs.mkShell {
          buildInputs = with pkgs; [ go git ];
        };
      });
}
