# Windows Local Installation & Setup Script for Multiout
# Usage: powershell -ExecutionPolicy Bypass -File .\scripts\install-windows.ps1

[CmdletBinding()]
param(
    [switch]$Build = $false,
    [switch]$NoDesktopShortcut = $false
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
Set-Location $projectRoot

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host " Multiout Windows Installer & Setup Script " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

$binDir = Join-Path $projectRoot "bin"
$exePath = Join-Path $binDir "multiout.exe"

# 1. Build binary if missing or requested
if ($Build -or !(Test-Path $exePath)) {
    Write-Host "`n[1/4] Building Multiout for Windows (GUI mode)..." -ForegroundColor Yellow
    if (!(Test-Path $binDir)) {
        New-Item -ItemType Directory -Path $binDir | Out-Null
    }
    go build -ldflags="-H=windowsgui -s -w" -o $exePath ./cmd/multiout
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed with exit code $LASTEXITCODE."
    }
    Write-Host "Built: $exePath" -ForegroundColor Green
} else {
    Write-Host "`n[1/4] Found existing binary: $exePath" -ForegroundColor Green
}

# 2. Destination folder
$installDir = Join-Path $env:LOCALAPPDATA "Programs\Multiout"
Write-Host "`n[2/4] Installing to $installDir..." -ForegroundColor Yellow
if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

Copy-Item -Path $exePath -Destination $installDir -Force
Copy-Item -Path (Join-Path $projectRoot "internal\ui\multiout-logo-rounded.png") -Destination $installDir -Force
Copy-Item -Path (Join-Path $projectRoot "README.md") -Destination $installDir -Force
if (Test-Path (Join-Path $projectRoot "LICENSE")) {
    Copy-Item -Path (Join-Path $projectRoot "LICENSE") -Destination $installDir -Force
}

$installedExe = Join-Path $installDir "multiout.exe"
Write-Host "Installed files successfully." -ForegroundColor Green

# 3. Create Shortcuts
Write-Host "`n[3/4] Creating application shortcuts..." -ForegroundColor Yellow
$wsh = New-Object -ComObject WScript.Shell

# Start Menu Shortcut
$startMenuDir = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs"
$startMenuShortcut = Join-Path $startMenuDir "Multiout.lnk"
$shortcut = $wsh.CreateShortcut($startMenuShortcut)
$shortcut.TargetPath = $installedExe
$shortcut.WorkingDirectory = $installDir
$shortcut.Description = "Simultaneous Multi-Device Audio Output"
$shortcut.Save()
Write-Host "Created Start Menu shortcut: $startMenuShortcut" -ForegroundColor Green

# Desktop Shortcut
if (-not $NoDesktopShortcut) {
    $desktopShortcut = Join-Path $env:USERPROFILE "Desktop\Multiout.lnk"
    $shortcut = $wsh.CreateShortcut($desktopShortcut)
    $shortcut.TargetPath = $installedExe
    $shortcut.WorkingDirectory = $installDir
    $shortcut.Description = "Simultaneous Multi-Device Audio Output"
    $shortcut.Save()
    Write-Host "Created Desktop shortcut: $desktopShortcut" -ForegroundColor Green
}

# 4. Check for Virtual Audio Cable
Write-Host "`n[4/4] Checking Virtual Audio Cable driver status..." -ForegroundColor Yellow
$hasCable = Get-CimInstance Win32_SoundDevice -ErrorAction SilentlyContinue | Where-Object {
    $_.Name -like "*VB-Audio*" -or $_.Name -like "*Virtual Cable*" -or $_.Name -like "*CABLE*"
}

if ($hasCable) {
    Write-Host "Virtual Audio Driver detected: $($hasCable[0].Name)" -ForegroundColor Green
    Write-Host "Multiout is ready for seamless system audio routing!" -ForegroundColor Green
} else {
    Write-Host "No Virtual Audio Cable detected." -ForegroundColor Yellow
    Write-Host "Tip: To route all Windows desktop sound without duplicate playback on your default speaker," -ForegroundColor Cyan
    Write-Host "     install VB-Audio Cable (free): https://vb-audio.com/Cable/" -ForegroundColor Cyan
}

Write-Host "`nInstallation complete! You can launch Multiout from the Desktop or Start Menu." -ForegroundColor Green
