Invoke-WebRequest "https://github.com/$env:GH_REPOSITORY/releases/download/$env:RELEASE_VERSION/phytofy-cli.windows-amd64.$env:RELEASE_VERSION.exe" -OutFile desktop\phytofy-cli.exe

npm i -g 'electron@11.1.1' 'electron-builder@22.9.1'
Copy-Item ui\public\img\icons\icon512.png desktop\icon.png
((Get-Content -path desktop\package.json -Raw) -replace '0.0.0',"$env:RELEASE_VERSION") | Set-Content -Path desktop\package.json
((Get-Content -path desktop\package-lock.json -Raw) -replace '0.0.0',"$env:RELEASE_VERSION") | Set-Content -Path desktop\package-lock.json

Set-Location desktop
npm ci
echo "HW v1"
npm run package-windows

choco install gh

echo $env:GH_API_TOKEN | gh auth login --hostname github.com --with-token
Rename-Item "release\OSRAM - PHYTOFY RL v1 UI Setup $env:RELEASE_VERSION.exe" "OSRAM - PHYTOFY RL v1 UI-$env:RELEASE_VERSION.exe"
gh release upload $env:RELEASE_VERSION "release/OSRAM - PHYTOFY RL v1 UI-$env:RELEASE_VERSION.exe"

Remove-Item -Recurse -Force release
((Get-Content -path desktop\package.json -Raw) -replace 'v1','v0') | Set-Content -Path desktop\package.json
((Get-Content -path desktop\package-lock.json -Raw) -replace 'v1','v0') | Set-Content -Path desktop\package-lock.json
((Get-Content -path desktop\main.json -Raw) -replace 'v1','v0') | Set-Content -Path desktop\main.js
echo "HW v0"
npm run package-windows

Rename-Item "release\OSRAM - PHYTOFY RL v0 UI Setup $env:RELEASE_VERSION.exe" "OSRAM - PHYTOFY RL v0 UI-$env:RELEASE_VERSION.exe"
gh release upload $env:RELEASE_VERSION "release/OSRAM - PHYTOFY RL v0 UI-$env:RELEASE_VERSION.exe"
