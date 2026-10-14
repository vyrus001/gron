$ErrorActionPreference = 'Stop'

if (Get-OSArchitectureWidth -Compare 32) {
  throw 'gron-vyrus001 0.8.1 requires Windows x64.'
}

$toolsDir = Split-Path -Parent $MyInvocation.MyCommand.Definition

Install-ChocolateyZipPackage `
  -PackageName 'gron-vyrus001' `
  -Url 'https://github.com/vyrus001/gron/releases/download/v0.8.1/gron_0.8.1_windows_amd64.zip' `
  -UnzipLocation $toolsDir `
  -Checksum '9b8ad90f2a6459f20560bb1d18a1a8aeda52d5483a362a8f3192215f7238b250' `
  -ChecksumType 'sha256'
