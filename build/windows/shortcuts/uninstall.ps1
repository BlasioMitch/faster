# Removes the Start Menu entry and Desktop shortcut created by install.ps1.
# faster.exe itself and your data (%APPDATA%\faster) are left untouched.

$startMenuPath = Join-Path ([Environment]::GetFolderPath("Programs")) "Faster.lnk"
$desktopPath = Join-Path ([Environment]::GetFolderPath("Desktop")) "Faster.lnk"

Remove-Item -Path $startMenuPath -ErrorAction SilentlyContinue
Remove-Item -Path $desktopPath -ErrorAction SilentlyContinue

Write-Host "Faster's Start Menu and Desktop shortcuts have been removed."
