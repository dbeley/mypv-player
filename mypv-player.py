#!/usr/bin/env python3
from pyfzf.pyfzf import FzfPrompt
from pathlib import Path
import subprocess
import argparse


FOLDER = "~/Téléchargements/Musique"
MPV_CMD = "mpv --no-video --display-tags=''"


def parse_args():
    parser = argparse.ArgumentParser(
        prog="mypv-player", description="A very simple music player using mpv and fzf"
    )
    parser.add_argument(
        "folder", nargs=1, type=str, help="Folder containing your music"
    )
    return parser.parse_args()


def get_song(folder: str):
    p = Path(folder).expanduser().glob("**/*")
    files = [str(x) for x in p if x.is_file()]
    fzf = FzfPrompt()
    choice = fzf.prompt(choices=files)
    return choice[0]


def play(audio_file):
    cmd = MPV_CMD.split()
    # cmd.insert(1, audio_file)
    cmd.append(audio_file)
    try:
        subprocess.run(cmd, check=False)
    except FileNotFoundError:
        print("File not found")


def main():
    args = parse_args()
    while True:
        audio_file = get_song(FOLDER)
        play(audio_file)


if __name__ == "__main__":
    main()
