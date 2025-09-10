# Sailshard — Units Demo (g3n client + server)

Working **micro‑voxel demo** with correct **g3n** deps.

- **Server**: TCP on `:27015`, 20 Hz tick broadcast + echo.
- **Client**: g3n window, free‑look camera, FPS counter, and a micro‑voxel island (1 block = 16×16×16 units). Press **Space** to carve.

## Prerequisites

- Go **1.22+**
- g3n engine deps (OpenGL + OpenAL, etc.)

### Linux (Debian/Ubuntu)
```bash
sudo apt-get install xorg-dev libgl1-mesa-dev libopenal1 libopenal-dev libvorbis0a libvorbis-dev libvorbisfile3
```

### macOS (Homebrew)
```bash
brew install libvorbis openal-soft
```

### Windows
Install [OpenAL Soft](https://openal-soft.org/) (provides openal32.dll).

## Build & Run

### Terminal 1 (server)
```bash
cd server/cmd/sailshardd
go run .
```

### Terminal 2 (client)
```bash
cd client/cmd/sailshard
go run .
```

Controls: W/A/S/D to move, Q/E up/down, RMB drag to look, Space to carve.
