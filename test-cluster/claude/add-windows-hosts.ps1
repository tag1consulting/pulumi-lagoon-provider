# PowerShell script to add Lagoon hosts to Windows hosts file
# Run this from an Administrator PowerShell window

$hostsFile = "C:\Windows\System32\drivers\etc\hosts"
$wslIP = (wsl hostname -I).Split()[0].Trim()

$entries = @(
    "$wslIP api.lagoon.test",
    "$wslIP ui.lagoon.test",
    "$wslIP keycloak.lagoon.test",
    "$wslIP harbor.lagoon.test"
)

Write-Host "WSL2 IP Address: $wslIP" -ForegroundColor Green
Write-Host "`nChecking Windows hosts file..." -ForegroundColor Yellow

$hostsContent = Get-Content $hostsFile

foreach ($entry in $entries) {
    $hostname = $entry.Split()[1]

    # Remove any existing entries for this hostname
    $hostsContent = $hostsContent | Where-Object { $_ -notmatch "\s+$hostname(\s|$)" }
}

# Add new entries
$hostsContent += ""
$hostsContent += "# Lagoon test cluster (added $(Get-Date))"
$hostsContent += $entries

# Write back to file
Set-Content -Path $hostsFile -Value $hostsContent -Force

Write-Host "`nSuccessfully added Lagoon entries to Windows hosts file:" -ForegroundColor Green
foreach ($entry in $entries) {
    Write-Host "  $entry" -ForegroundColor Cyan
}
Write-Host "`nYou can now access Lagoon from Windows browser:" -ForegroundColor Yellow
Write-Host "  - UI: http://ui.lagoon.test" -ForegroundColor Cyan
Write-Host "  - API: http://api.lagoon.test/graphql" -ForegroundColor Cyan
Write-Host "  - Keycloak: http://keycloak.lagoon.test/auth" -ForegroundColor Cyan
Write-Host "  - Harbor: http://harbor.lagoon.test" -ForegroundColor Cyan
