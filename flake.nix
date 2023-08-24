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
          vendorSha256 = "sha256-7PswrA6eVste+nYpoq93o4Jnly5dW7y2J/dOYQVaAh4=";
        };
      };
      devShells.${system} = {
        default = pkgs.mkShell { buildInputs = with pkgs; [ go ]; };
      };

    };
}
