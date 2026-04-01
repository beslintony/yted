# Build Directory

The build directory is used to house all the build files and assets for your application.

The structure is:

* **bin** - Output directory for compiled binaries
* **linux** - Linux-specific files (install scripts, `.desktop` entry, `.deb` packaging)
* **scripts** - Helper build scripts (FFmpeg bundling, `.deb` package creation)
* **windows** - Windows-specific files (icons, manifest, NSIS installer)

## Linux

The `linux` directory contains the installation script, desktop entry, and Debian packaging assets used when building and installing on Linux.

- `install.sh` - Shell script that installs the binary, icon, and `.desktop` file locally or system-wide. It also copies bundled `ffmpeg` and `ffprobe` if they exist.
- `yted.desktop` - Desktop entry for application launchers.

## Scripts

The `scripts` directory contains cross-platform helper scripts used during the build process.

- `bundle-ffmpeg.sh` - Downloads and bundles FFmpeg + FFprobe for Linux builds.
- `bundle-ffmpeg.ps1` - Downloads and bundles FFmpeg + FFprobe for Windows builds.
- `build-deb.sh` - Creates a `.deb` package from the compiled binary and bundled FFmpeg binaries.

These scripts are invoked automatically by the Makefile when running `make build`, `make build-installer-linux`, or `make build-installer-windows`.

## Windows

The `windows` directory contains the manifest, rc files, and NSIS installer files used when building with `wails build`.
These may be customised for your application. To return these files to the default state, simply delete them and
build with `wails build`.

- `icon.ico` - The icon used for the application. This is used when building using `wails build`. If you wish to
  use a different icon, simply replace this file with your own. If it is missing, a new `icon.ico` file
  will be created using the `appicon.png` file in the build directory.
- `installer/*` - The files used to create the Windows installer. These are used when building using `wails build --nsis`.
  The `project.nsi` script has been updated to conditionally bundle `ffmpeg.exe` and `ffprobe.exe` when available.
- `info.json` - Application details used for Windows builds. The data here will be used by the Windows installer,
  as well as the application itself (right click the exe -> properties -> details)
- `wails.exe.manifest` - The main application manifest file.
