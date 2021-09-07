##Built in PowerShell v5.1, but should work in any version.
$CUR_DIR = Get-Location | Select-Object -ExpandProperty Path
$SCRIPT_DIR = $PSScriptRoot
$date = [DateTimeOffset]::Now.ToUnixTimeSeconds()
Set-Location $SCRIPT_DIR
go build -o $CUR_DIR\op.exe -ldflags "-X main.gitCommit=$(git rev-parse HEAD) -X main.buildEpochSec=$date" "$SCRIPT_DIR\\cli\\op.go"
if ($?) {
    Set-Location $CUR_DIR
    $opargs = './op.exe ' + $args -join " "
    Invoke-Expression $opargs
} else {
    Write-Host "Failed to build op.exe. cli.ps1 may not be working correctly."
    Exit
}
