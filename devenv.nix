{ pkgs, lib, config, inputs, ... }:

{
  packages = [
    pkgs.git
    pkgs.tinygo
  ];

   languages.go.enable = true;

  scripts.hello.exec = ''
    echo hello from $GREET
  '';

  enterShell = ''
    hello
    git --version
  '';

}
