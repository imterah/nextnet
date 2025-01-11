{
  pkgs ? import <nixpkgs> { },
}: pkgs.mkShell {
  buildInputs = with pkgs; [
    # backend/
    nodejs
    go
    gopls
    capnproto
  ];

  shellHook = ''
    if [ -f init.sh ]; then
      source init.sh
    fi
  '';
}
