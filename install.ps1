#requires -Version 5.1
param(
 [string]$Distro = "Ubuntu",
 [ValidateSet("auto","off","internal")][string]$Tls = "off",
 [string]$Domain = "",
 [string]$AcmeEmail = "",
 [switch]$NonInteractive,
 [switch]$DryRun
)
$ErrorActionPreference = "Stop"
if (-not (Get-Command wsl.exe -ErrorAction SilentlyContinue)) {
 throw "WSL2 is required. Run wsl --install -d Ubuntu as Administrator, reboot if requested, then rerun this installer."
}
$Source = Split-Path -Parent $PSCommandPath
$LinuxPath = (& wsl.exe -d $Distro --exec wslpath -a $Source | Out-String).Trim()
if ($LASTEXITCODE -ne 0 -or -not $LinuxPath) { throw "WSL distribution is not initialized. Run wsl --install -d Ubuntu and complete its setup first." }
$LinuxArgs = @("-d",$Distro,"-u","root","--exec","bash",($LinuxPath+"/install.sh"),"--tls",$Tls)
if ($Domain) { $LinuxArgs += @("--domain",$Domain) }
if ($AcmeEmail) { $LinuxArgs += @("--email",$AcmeEmail) }
if ($NonInteractive) { $LinuxArgs += "--non-interactive" }
if ($DryRun) { $LinuxArgs += "--dry-run" }
& wsl.exe @LinuxArgs
if ($LASTEXITCODE -ne 0) { throw "Linux installer failed with code $LASTEXITCODE. Review the report; Windows installation is not reported successful." }
