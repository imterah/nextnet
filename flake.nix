{
  description = "custom flake template using flake parts";
  outputs = inputs:
    inputs.flake-parts.lib.mkFlake {inherit inputs;} {
      imports = [
        inputs.devshell.flakeModule
      ];
      systems = ["x86_64-linux"];
      perSystem = {
        pkgs,
        lib,
        self',
        inputs',
        ...
      }: {
        devshells.default = {
          packages = with pkgs; [
            # gui/
            cargo
            rustc
            gcc

            # api/
            nodejs
            openssl
            postgresql
            lsof
          ];
          commands = [
            {
              help = "start the api in dev mode.";
              name = "api.dev";
              command = ''cd "$(git rev-parse --show-toplevel)/api" && npm run dev &>api.log'';
              category = "api";
            }
            {
              help = "kill the api.";
              name = "api.kill";
              command = "kill -9 $(lsof -i :3000 | awk '{l=$2} END {print l}')";
              category = "api";
            }
            {
              help = "view the api log.";
              name = "api.log";
              command = ''tail -f "$(git rev-parse --show-toplevel)/api/api.log"'';
              category = "api";
            }
            {
              help = "start the gui.";
              name = "gui.run";
              command = ''cd "$(git rev-parse --show-toplevel)/gui" && cargo run'';
              category = "gui";
            }
          ];
          env = [];
          motd = ''
            {63}welcome to the nextnet devshell{reset}
            $(type -p menu &>/dev/null && menu)'';
          devshell.startup.default = lib.noDepEntry ''
            export PRISMA_QUERY_ENGINE_BINARY=${pkgs.prisma-engines}/bin/query-engine
            export PRISMA_QUERY_ENGINE_LIBRARY=${pkgs.prisma-engines}/lib/libquery_engine.node
            export PRISMA_SCHEMA_ENGINE_BINARY=${pkgs.prisma-engines}/bin/schema-engine

            # egui patches!
            export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${pkgs.wayland}/lib
            export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${pkgs.libxkbcommon}/lib
            export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${pkgs.xorg.libX11}/lib
            export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${pkgs.xorg.libXcursor}/lib
            export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${pkgs.xorg.libXrandr}/lib
            export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${pkgs.xorg.libXi}/lib
            export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${pkgs.libGL}/lib

            alias api.dev='$(cd "$(git rev-parse --show-toplevel)/api" && npm run dev &>api.log) & disown; echo "API started in the background."';

            source init.sh
          '';
        };
      };
    };
  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    devshell.url = "github:numtide/devshell";
  };
}
