# Proje araçlarını yalnızca bu terminal oturumu için etkinleştirir.
# Kullanım:  . .\env.ps1   (başındaki nokta önemli)
# Sistem PATH'ine, kayıt defterine veya kullanıcı profiline hiçbir şey yazılmaz.

$tools = Join-Path $PSScriptRoot '.tools'

# Araçlar eksikse ortamı hiç değiştirmeden, neyin nereye açılacağını söyleyip dur
$required = @(
    @{ Name = 'Go 1.27.3+';     Exe = 'go\bin\go.exe'; Dir = 'go';   Url = 'https://go.dev/dl/  (go1.27.3.windows-amd64.zip veya daha yenisi)' },
    @{ Name = 'Node.js 22.13+'; Exe = 'node\node.exe'; Dir = 'node'; Url = 'https://nodejs.org/en/download  (Windows .zip, 24 LTS önerilir)' }
)
$missing = @($required | Where-Object { -not (Test-Path (Join-Path $tools $_.Exe)) })
if ($missing.Count -gt 0) {
    Write-Host "Duplicate File Finder ortamı etkinleştirilemedi: gerekli araçlar bulunamadı." -ForegroundColor Red
    foreach ($m in $missing) {
        Write-Host ""
        Write-Host "  Eksik: $($m.Name)" -ForegroundColor Yellow
        Write-Host "    Beklenen dosya: $(Join-Path $tools $m.Exe)"
        Write-Host "    İndirme:        $($m.Url)"
        Write-Host "    Zip içeriğini $(Join-Path $tools $m.Dir) klasörüne açın."
    }
    Write-Host ""
    Write-Host "İsteğe bağlı Git: https://git-scm.com/download/win  (PortableGit, .tools\git klasörüne)"
    Write-Host "Ardından: . .\env.ps1  ve  go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.0"
    return
}

# Go/npm'in kullanıcı profiline yazdığı ayar ve telemetri dosyaları da .tools içinde kalsın
$env:APPDATA      = Join-Path $tools 'appdata'
New-Item -ItemType Directory -Force $env:APPDATA | Out-Null

$env:GOROOT       = Join-Path $tools 'go'
$env:GOPATH       = Join-Path $tools 'gopath'
$env:GOCACHE      = Join-Path $tools 'cache\go-build'
$env:GOMODCACHE   = Join-Path $tools 'gopath\pkg\mod'
$env:GOTOOLCHAIN  = 'local'
$env:npm_config_cache  = Join-Path $tools 'cache\npm'
$env:npm_config_prefix = Join-Path $tools 'node'

$paths = @(
    (Join-Path $env:GOROOT 'bin'),
    (Join-Path $env:GOPATH 'bin'),
    (Join-Path $tools 'node'),
    (Join-Path $tools 'git\cmd')
)
$env:PATH = ($paths -join ';') + ';' + $env:PATH

Write-Host "Duplicate File Finder ortamı etkin: $(go version), node $(node --version)" -ForegroundColor Green