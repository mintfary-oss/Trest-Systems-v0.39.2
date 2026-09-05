# Unified Windows/WSL launcher. Historical standalone installer is in archive/.
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
& (Join-Path $Root "install.ps1") @args
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
