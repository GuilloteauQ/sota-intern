{
  description = "A very basic flake";

  inputs = { nixpkgs.url = "github:nixos/nixpkgs/23.05"; };

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
    in {

      packages.${system} = rec {
        default = sota-intern;
        sota-intern = pkgs.buildGoModule {
          name = "sota-intern";
          version = "0.0";
          src = ./sota-intern;
          vendorSha256 = "sha256-7cJ7ikshlifKq8uW6prt2pE0W+V6N07EzXlHBHAXZjY=";
        };
      };
      devShells.${system} = {
        default = pkgs.mkShell { buildInputs = with pkgs; [ go ]; };
        record = pkgs.mkShell { buildInputs = with pkgs; [ vhs self.packages.${system}.sota-intern ]; };
      };

    };
}
