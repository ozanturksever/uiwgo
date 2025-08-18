{ pkgs, lib, config, inputs, ... }:

{
  packages = [
    pkgs.git
    pkgs.tinygo
  ];

   languages.go.enable = true;

  scripts = {
   "dev-build-deps" = {
  description = "build cli deps (use --force to rebuild all)";
  exec = ''
  FORCE=0
  if [ "$1" = "--force" ]; then
    FORCE=1
  fi
     if [ ! -f "$DEVENV_STATE/go/bin/go_js_wasm_exec" ] || [ $FORCE -eq 1 ]; then
       echo "Building deps..."

       go install github.com/agnivade/wasmbrowsertest@latest
       mv $DEVENV_STATE/go/bin/wasmbrowsertest $DEVENV_STATE/go/bin/go_js_wasm_exec
    fi
  '';
   };
  };

  enterShell = ''
    export PATH="$PATH:$DEVENV_ROOT/node_modules/.bin/:$DEVENV_STATE/go/bin/"

    dev-build-deps

    ${pkgs.gnused}/bin/sed -e 's| |••|g' -e 's|=| |' <<EOF | ${pkgs.util-linuxMinimal}/bin/column -t | ${pkgs.gnused}/bin/sed -e 's|^|🦾 |' -e 's|••| |g'
    ${lib.generators.toKeyValue { } (lib.mapAttrs (name: value: value.description) config.tasks)}
    EOF
    echo
    ${pkgs.gnused}/bin/sed -e 's| |••|g' -e 's|=| |' <<EOF | ${pkgs.util-linuxMinimal}/bin/column -t | ${pkgs.gnused}/bin/sed -e 's|^|🦾 |' -e 's|••| |g'
    ${lib.generators.toKeyValue { } (lib.mapAttrs (name: value: value.description) config.scripts)}
    EOF
    echo
  '';

}
