{
  pkgs ? import <nixpkgs> { },
}: pkgs.mkShell {
  buildInputs = with pkgs; [
    # api/
    nodejs
    go
    gopls
  ];

  shellHook = ''
    source init.sh
  '';
}
