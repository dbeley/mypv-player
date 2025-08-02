# mypv-player

`mypv-player` is a terminal user interface YouTube player written in Go.
It relies on `yt-dlp` and `mpv` to fetch and play videos.

Subscriptions are stored in `~/.config/mypv-player/config.json` and can be
managed from within the interface.

## Usage

```
go run .
```

### Controls
- Type a query and press **Enter** to search YouTube.
- **Enter** on a result plays the video.
- Press **a** on a result to add it to the queue.
- **Ctrl+q** shows the current queue.
- **Ctrl+s** shows latest videos from subscriptions.
- **Ctrl+m** opens the subscription manager; **a** adds a channel URL.
- **Ctrl+t** toggles audio-only playback.
- **Esc** quits or returns to the previous screen.

The development environment can be bootstrapped with `nix-shell` and
`direnv`.
