let
  pkgs = import (fetchTarball ("channel:nixpkgs-unstable")) { };
in
pkgs.mkShell {
  buildInputs = [ pkgs.nodejs pkgs.sqlite ];
}