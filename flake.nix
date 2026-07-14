{
  description = "mygo - A Cargo-like build tool for Go projects";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    treefmt-nix.url = "github:numtide/treefmt-nix";
  };

  outputs =
    {
      self,
      nixpkgs,
      treefmt-nix,
    }:
    let
      systems = [
        "aarch64-linux"
        "x86_64-linux"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f system);
      treefmtEval = forAllSystems (
        system: treefmt-nix.lib.evalModule nixpkgs.legacyPackages.${system} ./treefmt.nix
      );
    in
    {
      formatter = forAllSystems (system: treefmtEval.${system}.config.build.wrapper);

      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.buildGoModule {
            pname = "mygo";
            version = "git";
            src = ./.;
            subPackages = [ "cmd/mygo" ];
            vendorHash = "sha256-li+gf7hgCja2AAFBCZqVVfeaql70KcruQ4mGh3d7mM8=";
            nativeBuildInputs = [ pkgs.installShellFiles ];
            postInstall = ''
              installShellCompletion --cmd mygo \
                --bash <($out/bin/mygo completion bash) \
                --fish <($out/bin/mygo completion fish) \
                --zsh <($out/bin/mygo completion zsh)
            '';
          };
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              git
              treefmt-nix.packages.${system}.default
            ];
            shellHook = ''
              echo "mygo dev shell — go $(go version | cut -d' ' -f3)"
            '';
          };
        }
      );
    };
}
