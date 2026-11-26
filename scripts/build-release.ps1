$ErrorActionPreference = "Stop"

$RootDir = Split-Path -Parent $PSScriptRoot
Set-Location $RootDir

$env:PATH = "C:\Program Files\Go\bin;C:\Program Files\nodejs;$env:USERPROFILE\go\bin;" + $env:PATH
$env:CGO_ENABLED = "0"

Write-Host "==> [1/7] Verificando binarios FFmpeg / FFprobe..." -ForegroundColor Cyan
powershell -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot "fetch-ffmpeg.ps1")
if ($LASTEXITCODE -ne 0) { throw "fetch-ffmpeg.ps1 fallo con codigo $LASTEXITCODE" }

Write-Host "==> [2/7] Compilando frontend React + TypeScript..." -ForegroundColor Cyan
Push-Location frontend
try {
    npm.cmd run build
    if ($LASTEXITCODE -ne 0) { throw "npm.cmd run build fallo con codigo $LASTEXITCODE" }
} finally {
    Pop-Location
}

Write-Host "==> [3/7] Ejecutando suite de tests Go..." -ForegroundColor Cyan
go test -count=1 ./...
if ($LASTEXITCODE -ne 0) { throw "go test fallo con codigo $LASTEXITCODE" }

Write-Host "==> [4/7] Compilando binario nativo MediaPulse.exe (CGO=0, sin UPX)..." -ForegroundColor Cyan
wails build -clean -s -ldflags "-s -w"
if ($LASTEXITCODE -ne 0) { throw "wails build fallo con codigo $LASTEXITCODE" }

$ExePath = Join-Path $RootDir "build\bin\MediaPulse.exe"
if (-not (Test-Path $ExePath)) { throw "No se encontro $ExePath" }

$BinOutDir = Join-Path $RootDir "build\bin\bin"
New-Item -ItemType Directory -Force -Path $BinOutDir | Out-Null
Copy-Item (Join-Path $RootDir "build\ffmpeg\ffmpeg.exe") (Join-Path $BinOutDir "ffmpeg.exe") -Force
Copy-Item (Join-Path $RootDir "build\ffmpeg\ffprobe.exe") (Join-Path $BinOutDir "ffprobe.exe") -Force

$ReleaseDir = Join-Path $RootDir "release"
New-Item -ItemType Directory -Force -Path $ReleaseDir | Out-Null

$Version = "1.0.0"

$ReadmeText = @"
MediaPulse v$Version — Portable & Installer Windows x64
=======================================================

100% Local-First · 0% Telemetría · FFmpeg incluido · Aceleración GPU NVENC.

USO RÁPIDO:
- Ejecutá MediaPulse.exe directamente (incluye bin\ffmpeg.exe y bin\ffprobe.exe).
- Arrastrá tus videos o hacé clic en "Agregar archivos".
- Presets integrados:
  * Discord 10 MB / 25 MB: calcula el bitrate exacto para entrar debajo del límite.
  * WhatsApp: MP4 H.264 Main + yuv420p + AAC + faststart (cero errores).
  * TV Smart (MKV -> MP4): remux instantáneo (~2s) sin pérdida cuando ya es H.264.
  * Tamaño Exacto, Comprimir HQ, Extraer Audio (MP3/M4A/WAV) y Manual.
- Extras: Recorte rápido (Inicio/Fin), resolución, FPS, rotación y Mute.

AVISO DE WINDOWS SMARTSCREEN (Ejecutable Comunitario sin Firma EV):
Si Windows muestra "Windows protegió su PC", hacé clic en "Más información" -> "Ejecutar de todas formas".

VERIFICAR INTEGRIDAD (PowerShell):
   Get-FileHash .\MediaPulse.exe -Algorithm SHA256
   Compará el resultado con SHA256SUMS.txt.
"@

$ReadmePath = Join-Path $ReleaseDir "LEEME.txt"
Set-Content -Path $ReadmePath -Value $ReadmeText -Encoding UTF8

Write-Host "==> [5/7] Empaquetando ZIP portable..." -ForegroundColor Cyan
$StageDir = Join-Path $ReleaseDir "_stage_portable"
if (Test-Path $StageDir) { Remove-Item -Recurse -Force $StageDir }
New-Item -ItemType Directory -Force -Path $StageDir | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $StageDir "bin") | Out-Null
Copy-Item $ExePath (Join-Path $StageDir "MediaPulse.exe")
Copy-Item (Join-Path $BinOutDir "ffmpeg.exe") (Join-Path $StageDir "bin\ffmpeg.exe")
Copy-Item (Join-Path $BinOutDir "ffprobe.exe") (Join-Path $StageDir "bin\ffprobe.exe")
Copy-Item $ReadmePath (Join-Path $StageDir "LEEME.txt")

$ZipName = "MediaPulse-v$Version-win64-portable.zip"
$ZipPath = Join-Path $ReleaseDir $ZipName
if (Test-Path $ZipPath) { Remove-Item -Force $ZipPath }
Compress-Archive -Path (Join-Path $StageDir "*") -DestinationPath $ZipPath -CompressionLevel Optimal -Force
Remove-Item -Recurse -Force $StageDir

Write-Host "==> [6/7] Compilando instalador con Inno Setup 7..." -ForegroundColor Cyan
$ISCC = "C:\Program Files\Inno Setup 7\ISCC.exe"
if (-not (Test-Path $ISCC)) { throw "No se encontro ISCC.exe en $ISCC" }
$IssPath = Join-Path $RootDir "build\windows\installer\mediapulse-inno.iss"
& $ISCC "/DMyAppVersion=$Version" $IssPath
if ($LASTEXITCODE -ne 0) { throw "ISCC.exe fallo con codigo $LASTEXITCODE" }

$SetupName = "MediaPulse-v$Version-Setup.exe"
$SetupPath = Join-Path $ReleaseDir $SetupName
if (-not (Test-Path $SetupPath)) { throw "No se encontro $SetupPath" }

Write-Host "==> [7/7] Generando SHA256SUMS.txt..." -ForegroundColor Cyan
$Artifacts = @(
    @{ Path = $ExePath; Name = "MediaPulse.exe" },
    @{ Path = $ZipPath; Name = $ZipName },
    @{ Path = $SetupPath; Name = $SetupName }
)

$SumLines = @()
$HashByName = @{}
foreach ($a in $Artifacts) {
    $hash = (Get-FileHash -Path $a.Path -Algorithm SHA256).Hash.ToLowerInvariant()
    $SumLines += "$hash  $($a.Name)"
    $HashByName[$a.Name] = $hash
}
$SumsPath = Join-Path $ReleaseDir "SHA256SUMS.txt"
Set-Content -Path $SumsPath -Value $SumLines -Encoding ASCII

# Sync real SHA-256 checksums into WinGet & Scoop manifests
$WingetInstaller = Join-Path $RootDir "packaging\winget\MediaPulse.MediaPulse.installer.yaml"
if (Test-Path $WingetInstaller) {
    $setupHashUpper = $HashByName[$SetupName].ToUpperInvariant()
    $wingetContent = (Get-Content -Path $WingetInstaller -Raw) -replace 'InstallerSha256:\s*[A-Za-z0-9_]+', "InstallerSha256: $setupHashUpper"
    Set-Content -Path $WingetInstaller -Value $wingetContent.TrimEnd() -Encoding UTF8
}
$ScoopManifest = Join-Path $RootDir "packaging\scoop\mediapulse.json"
if (Test-Path $ScoopManifest) {
    $zipHashLower = $HashByName[$ZipName]
    $scoopContent = (Get-Content -Path $ScoopManifest -Raw) -replace '"hash":\s*"[^"]+"', "`"hash`": `"$zipHashLower`""
    Set-Content -Path $ScoopManifest -Value $scoopContent.TrimEnd() -Encoding UTF8
}

Write-Host "`n=== Release MediaPulse v$Version lista ===" -ForegroundColor Green
foreach ($a in $Artifacts) {
    $sizeMB = [math]::Round((Get-Item $a.Path).Length / 1MB, 2)
    Write-Host ("  - {0,-38} {1,6} MB" -f $a.Name, $sizeMB)
}
Write-Host "`nSHA256SUMS.txt:" -ForegroundColor Yellow
Get-Content $SumsPath
