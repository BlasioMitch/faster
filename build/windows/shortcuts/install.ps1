# Faster — adds a Start Menu entry and a Desktop shortcut for the current
# user. No admin rights needed; nothing outside your own profile is touched.
#
# Run it by right-clicking this file and choosing "Run with PowerShell",
# or from a PowerShell prompt in this folder:
#   powershell -ExecutionPolicy Bypass -File install.ps1

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$exePath = Join-Path $scriptDir "faster.exe"

if (-not (Test-Path $exePath)) {
    Write-Error "Could not find faster.exe next to this script (expected it in $scriptDir)."
    exit 1
}

$shell = New-Object -ComObject WScript.Shell

$startMenuPath = Join-Path ([Environment]::GetFolderPath("Programs")) "Faster.lnk"
$startMenuShortcut = $shell.CreateShortcut($startMenuPath)
$startMenuShortcut.TargetPath = $exePath
$startMenuShortcut.WorkingDirectory = $scriptDir
$startMenuShortcut.Description = "Faster - offline fasting tracker"
$startMenuShortcut.Save()

$desktopPath = Join-Path ([Environment]::GetFolderPath("Desktop")) "Faster.lnk"
$desktopShortcut = $shell.CreateShortcut($desktopPath)
$desktopShortcut.TargetPath = $exePath
$desktopShortcut.WorkingDirectory = $scriptDir
$desktopShortcut.Description = "Faster - offline fasting tracker"
$desktopShortcut.Save()

Write-Host "Installed. Faster now has a Start Menu entry and a Desktop shortcut."
Write-Host "To remove them later, run uninstall.ps1 from this same folder."
