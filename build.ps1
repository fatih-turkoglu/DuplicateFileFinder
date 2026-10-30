# Yayın derlemesi: build\bin\DuplicateFileFinder.exe
#   - sürüm wails.json'daki info.productVersion'dan alınır (exe özellikleri ve uygulama içi aynı)
#   - -trimpath: exe'ye derleme makinesinin yerel yolları gömülmez
#   - önce Go ve arayüz testleri çalışır; biri başarısızsa derleme yapılmaz
# Kullanım:  .\build.ps1            (testlerle)
#            .\build.ps1 -SkipTests
param([switch]$SkipTests)
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot
. .\env.ps1 | Out-Null

$version = (Get-Content wails.json -Raw | ConvertFrom-Json).info.productVersion
if (-not $version) { throw 'wails.json içinde info.productVersion yok' }

# Arayüz önce derlenir: main.go'daki //go:embed all:frontend/dist, go vet/test için
# frontend\dist'in var olmasını gerektirir (temiz klonda yoktur).
Push-Location frontend
try {
    npm ci --no-audit --no-fund
    if ($LASTEXITCODE) { throw 'npm ci başarısız' }
    npm run build
    if ($LASTEXITCODE) { throw 'arayüz derlemesi başarısız' }
} finally { Pop-Location }

if (-not $SkipTests) {
    go vet ./...
    if ($LASTEXITCODE) { throw 'go vet başarısız' }
    go test ./... -count=1
    if ($LASTEXITCODE) { throw 'Go testleri başarısız' }
    # Bilinen açık içeren Go sürümü ya da modülüyle derlenmesin.
    go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
    if ($LASTEXITCODE) { throw 'govulncheck açık buldu' }
    Push-Location frontend
    try {
        npm test
        if ($LASTEXITCODE) { throw 'arayüz testleri başarısız' }
    } finally { Pop-Location }
}

# Simge PNG'si değiştiyse .ico yeniden üretilsin (Wails yalnızca yoksa üretir).
$ico = 'build\windows\icon.ico'
if ((Test-Path $ico) -and (Get-Item 'build\appicon.png').LastWriteTime -gt (Get-Item $ico).LastWriteTime) {
    Remove-Item $ico
}

# -s: arayüz yukarıda derlendi, Wails yeniden derlemesin
wails build -clean -s -trimpath -ldflags "-X main.version=$version"
if ($LASTEXITCODE) { throw 'wails build başarısız' }

# Dil dosyaları exe'nin yanındaki lang klasörüne (uygulama oradan okur).
$langOut = 'build\bin\lang'
New-Item -ItemType Directory -Force $langOut | Out-Null
Copy-Item 'lang\*.lng' $langOut -Force
Write-Host ("Dil dosyaları: {0}" -f ((Get-ChildItem $langOut -Filter *.lng).BaseName -join ', '))

$exe = Get-Item 'build\bin\DuplicateFileFinder.exe'

# Dağıtım için Authenticode imzası (isteğe bağlı): DUPLICATEFILE_SIGN_CERT, kullanıcı
# sertifika deposundaki kod imzalama sertifikasının SHA-1 parmak izidir.
if ($env:DUPLICATEFILE_SIGN_CERT) {
    $signtool = (Get-Command signtool.exe -ErrorAction SilentlyContinue).Source
    if (-not $signtool) {
        $signtool = Get-ChildItem "${env:ProgramFiles(x86)}\Windows Kits\10\bin\*\x64\signtool.exe" -ErrorAction SilentlyContinue |
            Sort-Object FullName | Select-Object -Last 1 -ExpandProperty FullName
    }
    if (-not $signtool) { throw 'signtool.exe bulunamadı (Windows SDK gerekli)' }
    $tsa = if ($env:DUPLICATEFILE_TIMESTAMP_URL) { $env:DUPLICATEFILE_TIMESTAMP_URL } else { 'http://timestamp.digicert.com' }
    & $signtool sign /sha1 $env:DUPLICATEFILE_SIGN_CERT /fd SHA256 /tr $tsa /td SHA256 $exe.FullName
    if ($LASTEXITCODE) { throw 'imzalama başarısız' }
    & $signtool verify /pa $exe.FullName
    if ($LASTEXITCODE) { throw 'imza doğrulanamadı' }
} else {
    Write-Host 'İmzasız derleme (dağıtım için DUPLICATEFILE_SIGN_CERT tanımlayın).' -ForegroundColor Yellow
}

# İndirenlerin dosyayı doğrulayabilmesi için SHA-256 özetleri (sha256sum biçimi).
$sums = Get-ChildItem 'build\bin' -Recurse -File -Include *.exe, *.lng | ForEach-Object {
    '{0}  {1}' -f (Get-FileHash $_.FullName -Algorithm SHA256).Hash.ToLower(), $_.FullName.Substring((Resolve-Path 'build\bin').Path.Length + 1).Replace('\', '/')
}
[IO.File]::WriteAllLines((Join-Path (Resolve-Path 'build\bin') 'SHA256SUMS.txt'), [string[]]$sums)

# Yayın arşivi: exe + lang. Adda sürüm yok; README'deki releases/latest/download bağlantısı sabit kalsın.
$zip = 'build\bin\DuplicateFileFinder-windows-amd64.zip'
Remove-Item $zip -ErrorAction SilentlyContinue
Compress-Archive -Path $exe.FullName, $langOut -DestinationPath $zip
$zipHash = (Get-FileHash $zip -Algorithm SHA256).Hash.ToLower()
[IO.File]::WriteAllText((Resolve-Path $zip).Path + '.sha256', "$zipHash  $(Split-Path $zip -Leaf)`n")
Write-Host "Arşiv: $zip"

Write-Host ("Duplicate File Finder {0}: {1} ({2:N1} MB)" -f $version, $exe.FullName, ($exe.Length / 1MB)) -ForegroundColor Green
