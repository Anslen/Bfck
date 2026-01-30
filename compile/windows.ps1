$ErrorActionPreference = "Stop"

# Switch to script directory
Set-Location $PSScriptRoot

# Create bin directory if it doesn't exist
if (-not (Test-Path "../bin")) {
    New-Item -ItemType Directory -Path "../bin" | Out-Null
}

Write-Host "Building Bfck..."

# Build the project
Push-Location ../src
try {
    go build -ldflags "-s -w" -o ../bin/bfck.exe .
    Write-Host "Build successful! Output: bin\bfck.exe" -ForegroundColor Green
} catch {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
}
