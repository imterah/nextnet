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
    if [ -f init.sh ]; then
      source init.sh
    fi
  '';
}
