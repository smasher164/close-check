{
  description = "close-check";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let supportedSystems = [
      "aarch64-linux"
      "i686-linux"
      "x86_64-linux"
    ]; in
    flake-utils.lib.eachSystem supportedSystems (system:
      let
        vscode-overlay = final: prev: {
          vscodium = prev.vscodium.overrideAttrs (old: {
            buildInputs = old.buildInputs or [ ] ++ [ final.makeWrapper ];
            postFixup = old.postFixup or "" + ''
              wrapProgram $out/bin/${pkgs.vscodium.executableName} \
                --add-flags "--enable-features=UseOzonePlatform --ozone-platform=wayland"
            '';
          });
        };
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ vscode-overlay ];
        };
        vscode-with-extensions = pkgs.vscode-with-extensions.override {
          vscode = pkgs.vscodium;
          vscodeExtensions = [
            pkgs.vscode-extensions.golang.go
          ];
        };
      in
      {
        devShell = pkgs.mkShell {
          buildInputs = [
            pkgs.go_1_17
            pkgs.gopls
            vscode-with-extensions
          ];
        };
      });
}
