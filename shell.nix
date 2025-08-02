{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.mpv
    pkgs.yt-dlp
  ];
}
