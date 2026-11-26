param(
    [switch]$Force
)

$ErrorActionPreference = "Stop"

$RootDir = Split-Path -Parent $PSScriptRoot
$OutDir = Join-Path $RootDir "build\ffmpeg"
$FFmpegExe = Join-Path $OutDir "ffmpeg.exe"
$FFprobeExe = Join-Path $OutDir "ffprobe.exe"

if (-not $Force -and (Test-Path $FFmpegExe) -and (Test-Path $FFprobeExe)) {
    Write-Host "==> FFmpeg y FFprobe ya presentes en $OutDir" -ForegroundColor Green
    & $FFmpegExe -version | Select-Object -First 1
    exit 0
}

New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("mediapulse_ffmpeg_" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Force -Path $TempDir | Out-Null

$ZipPath = Join-Path $TempDir "ffmpeg.zip"
$ExtractDir = Join-Path $TempDir "extract"

$Urls = @(
    "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip",
    "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip"
)

$Downloaded = $false

foreach ($url in $Urls) {
    try {
        Write-Host "==> Descargando FFmpeg Windows x64 desde $url ..." -ForegroundColor Cyan
        $curl = Get-Command curl.exe -ErrorAction SilentlyContinue
        if ($curl) {
            & curl.exe -fL --connect-timeout 15 --max-time 180 --retry 1 -o $ZipPath $url
            if ($LASTEXITCODE -eq 0 -and (Test-Path $ZipPath) -and ((Get-Item $ZipPath).Length -gt 5MB)) {
                $Downloaded = $true
                break
            }
        } else {
            [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
            $ProgressPreference = 'SilentlyContinue'
            Invoke-WebRequest -Uri $url -OutFile $ZipPath -UseBasicParsing -TimeoutSec 180
            if ((Test-Path $ZipPath) -and ((Get-Item $ZipPath).Length -gt 5MB)) {
                $Downloaded = $true
                break
            }
        }
    } catch {
        Write-Warning "Fallo descarga desde $url : $($_.Exception.Message)"
    }
    if (Test-Path $ZipPath) { Remove-Item -Force $ZipPath -ErrorAction SilentlyContinue }
}

try {
    if (-not $Downloaded) {
        throw "No se pudo descargar FFmpeg desde ninguna fuente."
    }

    Write-Host "==> Extrayendo ffmpeg.exe y ffprobe.exe..." -ForegroundColor Cyan
    Expand-Archive -Path $ZipPath -DestinationPath $ExtractDir -Force

    $foundFFmpeg = Get-ChildItem -Path $ExtractDir -Recurse -Filter "ffmpeg.exe" | Select-Object -First 1
    $foundFFprobe = Get-ChildItem -Path $ExtractDir -Recurse -Filter "ffprobe.exe" | Select-Object -First 1

    if (-not $foundFFmpeg -or -not $foundFFprobe) {
        throw "El archivo comprimido no contiene ffmpeg.exe y ffprobe.exe."
    }

    Copy-Item $foundFFmpeg.FullName $FFmpegExe -Force
    Copy-Item $foundFFprobe.FullName $FFprobeExe -Force

    $verOut = & $FFmpegExe -version 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "ffmpeg.exe -version fallo con codigo $LASTEXITCODE"
    }

    Write-Host "==> FFmpeg listo:" -ForegroundColor Green
    Write-Host "    $($verOut | Select-Object -First 1)" -ForegroundColor Gray
    Write-Host "    ffmpeg.exe  : $([math]::Round((Get-Item $FFmpegExe).Length / 1MB, 2)) MB" -ForegroundColor Gray
    Write-Host "    ffprobe.exe : $([math]::Round((Get-Item $FFprobeExe).Length / 1MB, 2)) MB" -ForegroundColor Gray
} finally {
    if (Test-Path $TempDir) {
        Remove-Item -Recurse -Force $TempDir -ErrorAction SilentlyContinue
    }
}
