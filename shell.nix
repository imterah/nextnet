{
  pkgs ? import <nixpkgs> { },
}: pkgs.mkShell {
  buildInputs = with pkgs; [ 
    # gui/
    cargo
    rustc
    
    # api/
    nodejs
    openssl
    postgresql
    lsof
  ];  

  shellHook = ''
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

    if [ ! -d ".tmp" ]; then
      echo "Hello and welcome to the NextNet project!"
      mkdir .tmp
    fi
    
    source init.sh
  '';
}