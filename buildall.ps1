# buildall.ps1
Param(
    [string]$Platform = ""
)

# Pfad zu Wails (prüft ob es im PATH ist, sonst Standard-Go-Pfad)
$WAILS = "wails"
if (!(Get-Command $WAILS -ErrorAction SilentlyContinue)) {
    $WAILS = "$env:USERPROFILE\go\bin\wails.exe"
}

$buildArgs = @("build")
if ($Platform) {
    $buildArgs += "-platform"
    $buildArgs += $Platform
}

# Funktion zum Ändern des Namens in der wails.json
function Set-WailsName($newName) {
    $config = Get-Content wails.json | ConvertFrom-Json
    $config.name = $newName
    $config.outputfilename = $newName
    $config | ConvertTo-Json -Depth 10 | Set-Content wails.json
}

# Original-Namen sichern
if (Test-Path "wails.json") {
    $originalConfig = Get-Content wails.json | ConvertFrom-Json
    $origName = $originalConfig.name
} else {
    Write-Host "Fehler: wails.json nicht gefunden!" -ForegroundColor Red
    exit
}

Write-Host "------------------------------------------" -ForegroundColor Cyan
Write-Host "Baue HuhnLite-Local (SQLite)..."
Set-WailsName "HuhnLite-Local"
& $WAILS $buildArgs

if (Test-Path "settings.json") {
    if (!(Test-Path "build\bin")) { New-Item -ItemType Directory -Path "build\bin" -Force }
    Copy-Item "settings.json" "build\bin\settings.json" -Force
    Write-Host "settings.json kopiert." -ForegroundColor Gray
}

if (Test-Path "HuhnLite.db") {
    if (!(Test-Path "build\bin")) { New-Item -ItemType Directory -Path "build\bin" -Force }
    Copy-Item "HuhnLite.db" "build\bin\HuhnLite.db" -Force
    Write-Host "HuhnLite.db kopiert." -ForegroundColor Gray
}

Write-Host "------------------------------------------" -ForegroundColor Cyan
Write-Host "Baue HuhnLite-MariaDB (Netzwerk)..."
Set-WailsName "HuhnLite-MariaDB"
& $WAILS $buildArgs

if (Test-Path "settings_mariadb.json") {
    if (!(Test-Path "build\bin")) { New-Item -ItemType Directory -Path "build\bin" -Force }
    Copy-Item "settings_mariadb.json" "build\bin\settings_mariadb.json" -Force
    Write-Host "settings_mariadb.json kopiert." -ForegroundColor Gray
}

Write-Host "------------------------------------------" -ForegroundColor Cyan
Write-Host "Baue HuhnLite-Server (Client-Server SQLite)..."
$targetOS = "windows"
if ($Platform) {
    $parts = $Platform -split "/"
    $targetOS = $parts[0]
    $targetArch = $parts[1]
    $oldGOOS = $env:GOOS
    $oldGOARCH = $env:GOARCH
    $env:GOOS = $targetOS
    $env:GOARCH = $targetArch
}

$serverName = "HuhnLite-Server"
if ($targetOS -eq "windows") {
    $serverName = "HuhnLite-Server.exe"
}

go build -ldflags="-w -s" -o "build/bin/$serverName" .

if ($Platform) {
    $env:GOOS = $oldGOOS
    $env:GOARCH = $oldGOARCH
}

if (Test-Path "settings_server.json") {
    if (!(Test-Path "build\bin")) { New-Item -ItemType Directory -Path "build\bin" -Force }
    Copy-Item "settings_server.json" "build\bin\settings_server.json" -Force
    Write-Host "settings_server.json kopiert." -ForegroundColor Gray
}

Write-Host "------------------------------------------" -ForegroundColor Cyan
Write-Host "Baue HuhnLite-Server-MariaDB (Client-Server MariaDB)..."
if ($Platform) {
    $env:GOOS = $targetOS
    $env:GOARCH = $targetArch
}

$serverMariaDBName = "HuhnLite-Server-MariaDB"
if ($targetOS -eq "windows") {
    $serverMariaDBName = "HuhnLite-Server-MariaDB.exe"
}

go build -ldflags="-w -s" -o "build/bin/$serverMariaDBName" .

if ($Platform) {
    $env:GOOS = $oldGOOS
    $env:GOARCH = $oldGOARCH
}

if (Test-Path "settings_server_mariadb.json") {
    if (!(Test-Path "build\bin")) { New-Item -ItemType Directory -Path "build\bin" -Force }
    Copy-Item "settings_server_mariadb.json" "build\bin\settings_server_mariadb.json" -Force
    Write-Host "settings_server_mariadb.json kopiert." -ForegroundColor Gray
}

# Original-Zustand wiederherstellen
Set-WailsName $origName

Write-Host "------------------------------------------" -ForegroundColor Green
Write-Host "Fertig! Die Dateien wurden in build\bin\ bereitgestellt."
if (Test-Path "build\bin") {
    Get-ChildItem "build\bin\*.json"
}
