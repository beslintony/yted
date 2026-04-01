# Bundle FFmpeg binaries for Windows
$ErrorActionPreference = "Stop"

$binDir = "build\bin"
New-Item -ItemType Directory -Force -Path $binDir | Out-Null

if (Test-Path "$binDir\ffmpeg.exe") {
    Write-Host "FFmpeg already bundled"
    exit 0
}

Write-Host "Downloading FFmpeg for Windows..."
$zipPath = "$env:TEMP\ffmpeg-windows.zip"
Invoke-WebRequest -Uri "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip" -OutFile $zipPath

Expand-Archive -Path $zipPath -DestinationPath "$env:TEMP\ffmpeg-windows" -Force
$dir = Get-ChildItem -Path "$env:TEMP\ffmpeg-windows" -Filter "ffmpeg-*-essentials_build" | Select-Object -First 1

Copy-Item "$($dir.FullName)\bin\ffmpeg.exe" -Destination "$binDir\ffmpeg.exe" -Force
Copy-Item "$($dir.FullName)\bin\ffprobe.exe" -Destination "$binDir\ffprobe.exe" -Force

Remove-Item -Recurse -Force $zipPath, "$env:TEMP\ffmpeg-windows"
Write-Host "FFmpeg bundled for Windows"
