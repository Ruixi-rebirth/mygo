{
  description = "mygo - A Cargo-like build tool for Go projects";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    treefmt-nix.url = "github:numtide/treefmt-nix";
    pre-commit-hooks.url = "github:cachix/pre-commit-hooks.nix";
  };

  outputs =
    {
      self,
      nixpkgs,
      treefmt-nix,
      pre-commit-hooks,
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
            vendorHash = "sha256-7nE5Jkn/ItezXMQKHRKuPE5wztvPcDDeH4UF39OMyj0=";
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
          pre-commit-check = pre-commit-hooks.lib.${system}.run {
            src = ./.;
            hooks = {
              gofmt.enable = true;
              golangci-lint = {
                enable = true;
                entry = "${pkgs.golangci-lint}/bin/golangci-lint run --config ${./.golangci.yml}";
                files = "\\.go$";
                pass_filenames = false;
              };
              govet = {
                enable = true;
                entry = "${pkgs.go}/bin/go vet ./...";
                files = "\\.go$";
                pass_filenames = false;
              };
              nixfmt.enable = true;
            };
          };
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              git
              golangci-lint
            ];
            shellHook = ''
              echo "mygo — go $(go version | cut -d' ' -f3)"
            ''
            # Pre-commit hooks are installed by pre-commit-check.shellHook
            + pre-commit-check.shellHook;
          };
        }
      );
    };
}
