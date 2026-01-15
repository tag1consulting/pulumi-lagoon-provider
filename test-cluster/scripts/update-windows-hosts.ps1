# PowerShell script to update Windows hosts file with Lagoon domains
# Run this from PowerShell as Administrator on Windows

$hostsFile = "C:\Windows\System32\drivers\etc\hosts"
$marker = "# Lagoon test cluster"
$entries = "127.0.0.1 api.lagoon.test keycloak.lagoon.test ui.lagoon.test harbor.lagoon.test"

# Check if running as Administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Error "This script must be run as Administrator"
    Write-Host "Right-click PowerShell and select 'Run as Administrator', then run this script again"
    exit 1
}

# Check if entries already exist
$content = Get-Content $hostsFile -Raw
if ($content -match [regex]::Escape($marker)) {
    Write-Host "Lagoon hosts entries already exist in $hostsFile"
    exit 0
}

# Add entries
Write-Host "Adding Lagoon hosts entries to $hostsFile"
Add-Content -Path $hostsFile -Value "`n$marker"
Add-Content -Path $hostsFile -Value $entries

Write-Host "Done! Added:"
Write-Host "  $entries"
Write-Host ""
Write-Host "You can now access Lagoon services from Windows browsers at:"
Write-Host "  http://api.lagoon.test/graphql"
Write-Host "  http://keycloak.lagoon.test/auth"
Write-Host "  http://ui.lagoon.test"
Write-Host "  http://harbor.lagoon.test"
