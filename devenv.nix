{ pkgs, lib, config, inputs, ... }:

{
  packages = [
    pkgs.git
    pkgs.tinygo
  ];

   languages.go.enable = true;

  scripts = {
      setup-sdk = {
          exec = ''
     echo "Setting deterministic SDK dirs"
     if [ ! -f "$DEVENV_STATE/go/VERSION" ] || ! diff "$GOROOT/VERSION" "$DEVENV_STATE/go/VERSION" >/dev/null 2>&1; then
       rm -rf $DEVENV_STATE/go
       cp -r $GOROOT $DEVENV_STATE
       chmod -R 755 $DEVENV_STATE/go

       mv $DEVENV_STATE/go/bin/go $DEVENV_STATE/go/bin/go.orig
       go get github.com/ozanturksever/gowrapper@latest
       go install github.com/ozanturksever/gowrapper@latest
       mv $DEVENV_STATE/go/bin/gowrapper $DEVENV_STATE/go/bin/go
       chmod +x $DEVENV_STATE/go/bin/go
       echo "GoSDK(updated) at $DEVENV_STATE/go/"
     else
       echo "GoSDK(current) at $DEVENV_STATE/go/"
     fi
   '';
        };
   "dev-build-deps " = {
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

    setup-sdk # creates state/go
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
