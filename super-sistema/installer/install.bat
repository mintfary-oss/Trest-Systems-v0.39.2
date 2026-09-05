@echo off
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0..\..\install.ps1" %*
exit /b %ERRORLEVEL%
