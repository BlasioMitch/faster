Faster for Windows
===================

1. Run install.ps1 to add a Start Menu entry and a Desktop shortcut:

   - Easiest: right-click install.ps1 -> "Run with PowerShell".

   - If Windows blocks it with a script-execution-policy message, open
     PowerShell in this folder instead and run:

       powershell -ExecutionPolicy Bypass -File install.ps1

2. Launch Faster from the Start Menu or Desktop shortcut, or just run
   faster.exe directly — installing the shortcuts is optional.

3. To remove the shortcuts later, run uninstall.ps1 the same way.
   Your data (in %APPDATA%\faster) is never touched by either script.
