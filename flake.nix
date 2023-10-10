{
  description = "A very basic flake";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      shell = { pkgs }: pkgs.mkShell {
        buildInputs = with pkgs; [
          go_1_21
          gopls
          pprof
          graphviz
        ];
        hardeningDisable = [ "fortify" ];
      };
    in
    {
      devShell.x86_64-linux = shell { pkgs = nixpkgs.legacyPackages.x86_64-linux; };
      devShell.aarch64-darwin = shell { pkgs = nixpkgs.legacyPackages.aarch64-darwin; };
    };
}
