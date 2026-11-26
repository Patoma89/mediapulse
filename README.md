# MediaPulse v1.0.0

[![Release](https://img.shields.io/github/v/release/Patoma89/mediapulse?style=flat-square&color=8B5CF6)](https://github.com/Patoma89/mediapulse/releases/latest)
[![CI](https://github.com/Patoma89/mediapulse/actions/workflows/ci.yml/badge.svg)](https://github.com/Patoma89/mediapulse/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-10B981?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Windows-10%20%2F%2011%20x64-0EA5E9?style=flat-square)]()
[![Zero Telemetry](https://img.shields.io/badge/Telemetry-0%25%20Local--First-10B981?style=flat-square)]()
[![GPU](https://img.shields.io/badge/Hardware-NVENC%20%2F%20AMF%20%2F%20QSV-8B5CF6?style=flat-square)]()

> **Web oficial & Descargas:** https://patoma89.github.io/mediapulse/  
> **Último Release:** https://github.com/Patoma89/mediapulse/releases/latest

Conversor de video y audio ultrarrápido y **100% local-first** para **Windows 10 / 11 (x64)** de la familia **Patoma89** (mismo stack nativo Go + Wails v2 + React que [NetPulse](https://github.com/Patoma89/netpulse)). Diseñado para resolver en 1 clic los casos reales de uso diario sin lidiar con flags de FFmpeg: **Discord (10/25 MB)**, **WhatsApp**, **Smart TVs (remux MKV → MP4 en segundos)**, tamaño objetivo exacto, recorte rápido y extracción de audio.

### Descarga Rápida (`v1.0.0`)

| Opción | Archivo / Comando | Descripción |
| :--- | :--- | :--- |
| **Instalador Windows** | [`MediaPulse-v1.0.0-Setup.exe`](https://github.com/Patoma89/mediapulse/releases/download/v1.0.0/MediaPulse-v1.0.0-Setup.exe) | Instalador Inno Setup con FFmpeg incluido (usuario estándar, sin Admin). |
| **Portable (.zip)** | [`MediaPulse-v1.0.0-win64-portable.zip`](https://github.com/Patoma89/mediapulse/releases/download/v1.0.0/MediaPulse-v1.0.0-win64-portable.zip) | Carpeta portable lista para usar (`MediaPulse.exe` + `bin\ffmpeg.exe` + `LEEME.txt`). |
| **Checksums** | [`SHA256SUMS.txt`](https://github.com/Patoma89/mediapulse/releases/download/v1.0.0/SHA256SUMS.txt) | Huellas SHA-256 verificables en PowerShell. |

---

## Características Principales

- **FFmpeg & FFprobe 100% incluidos:** tanto el instalador como el ZIP portable ya traen `bin\ffmpeg.exe` y `bin\ffprobe.exe` listos para usar offline, sin descargas al abrir.
- **Aceleración por Hardware Automática:** prueba en vivo **NVIDIA NVENC (`h264_nvenc` / `hevc_nvenc`)** (ideal para RTX 3090 / Turing+), **AMD AMF**, **Intel QuickSync (QSV)** y fallback automático a **CPU (`libx264` / `libx265`)** si un encoder hardware falla.
- **Presets Inteligentes en 1 Clic:**
  - **Discord 10 MB / 25 MB:** calcula el bitrate exacto (con margen de seguridad `0.95` para overhead MP4 y `-multipass fullres` en NVENC o 2 pasadas reales en CPU) para quedar siempre por debajo del límite.
  - **WhatsApp:** fuerza `MP4 + H.264 Main + yuv420p + AAC + faststart` para evitar videos en negro o errores al enviar.
  - **TV Smart (`MKV → MP4`):** analiza los streams con `ffprobe`:
    - Si el MKV ya tiene H.264 8-bit y audio compatible → **Remux instantáneo (~2s)** sin recodificar ni perder calidad.
    - Si tiene audio incompatible (DTS/AC3/Opus) → copia el video intacto y convierte solo el audio a AAC.
    - Si es HEVC/AV1/10-bit → recodifica a H.264 universal para TVs.
  - **Tamaño Exacto (MB):** elegís el peso objetivo (ej. 50 MB) y calcula todo automáticamente.
  - **Comprimir HQ / Manual:** control de formato (`MP4`, `MKV`, `MOV`, `WEBM`), códec (`H.264`, `HEVC`, `Copy`) y calidad CRF/CQ.
  - **Extraer Audio:** salida directa en `MP3`, `M4A (AAC)` o `WAV` sin recodificar video.
- **Edición Rápida sin Complicaciones:** recorte Inicio/Fin (`MM:SS` o segundos), cambio de resolución (`1080p`, `720p`, `480p`), FPS (`60`, `30`, `24`), rotación (`90°`, `180°`, `270°`) y **Quitar Audio (Mute)**.
- **Cola por Lotes + Arrastrar y Soltar:** procesa múltiples archivos en secuencia con progreso en vivo (`%`, velocidad `x`, ETA), escritura atómica temporal y registro local de ahorro de espacio en `%APPDATA%\MediaPulse`.
- **0% Telemetría, ES/EN y Tema Oscuro/Claro.**
- Manifiestos **WinGet** (`packaging/winget/`) y **Scoop** (`packaging/scoop/`).

---

## Aviso de Windows SmartScreen

MediaPulse es una utilidad comunitaria gratuita sin certificado comercial EV de Microsoft. La primera vez que la ejecutes, Windows Defender SmartScreen puede mostrar *"Windows protegió su PC"*:
1. Hacé clic en **Más información**.
2. Hacé clic en **Ejecutar de todas formas**.

Verificá la huella SHA-256 contra `release/SHA256SUMS.txt`:
```powershell
Get-FileHash .\MediaPulse.exe -Algorithm SHA256
```

---

## Compilar desde el Código Fuente

Requisitos en Windows 10/11 x64: **Go 1.26+**, **Node.js LTS**, **Wails CLI v2** y **Inno Setup 7** (para generar `Setup.exe`).

```powershell
$env:CGO_ENABLED = "0"
powershell -ExecutionPolicy Bypass -File .\scripts\build-release.ps1
```

Los artefactos finales se generan en `release/`:
- `MediaPulse-v1.0.0-win64-portable.zip`
- `MediaPulse-v1.0.0-Setup.exe`
- `SHA256SUMS.txt`
