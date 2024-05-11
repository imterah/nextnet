{
  pkgs ? import <nixpkgs> { },
}: pkgs.mkShell {
  buildInputs = with pkgs; [ 
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

    if [ ! -d ".tmp" ]; then
      echo "Hello and welcome to the NextNet project!"
      mkdir .tmp
    fi
    
    source init.sh
  '';
}
