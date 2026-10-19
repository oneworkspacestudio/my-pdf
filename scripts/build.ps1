param([string]$Go='go',[string]$Npm='npm.cmd',[string]$Version='0.1.0-preview.1',[string]$Output='', [switch]$ReuseDependencies,[switch]$Release)
$ErrorActionPreference='Stop'
$root=Split-Path $PSScriptRoot -Parent
$contract=Get-Content -LiteralPath (Join-Path $root 'compatibility.json') -Raw|ConvertFrom-Json
$repo=$contract.repository;$id=$contract.app;$exe=$contract.executable
if($Version -notmatch '^\d+\.\d+\.\d+(-[A-Za-z0-9.]+)?$'){throw 'Invalid release version'}
if(-not $Output){$Output=Join-Path $root 'dist'}
$Output=[IO.Path]::GetFullPath($Output)
New-Item -ItemType Directory -Force -Path $Output|Out-Null
$env:GOWORK='off';$env:CGO_ENABLED='0';$env:GOOS='windows';$env:GOARCH='amd64'
Push-Location $root
try{
 $safe='safe.directory='+$root.Replace('\','/')
 $source=(& git -c $safe rev-parse HEAD).Trim()
 if($Release){if($LASTEXITCODE -or $source -notmatch '^[a-f0-9]{40}$'){throw 'Commit release source before packaging'};if(& git -c $safe status --porcelain){throw 'Release builds require a clean committed source tree'}}
 Push-Location (Join-Path $root 'web')
 try{
  if(-not $ReuseDependencies){& $Npm ci --audit=false --fund=false;if($LASTEXITCODE){throw 'npm ci failed'}}
  & $Npm run build;if($LASTEXITCODE){throw 'Frontend build failed'}
  & $Npm test;if($LASTEXITCODE){throw 'Frontend contract tests failed'}
 }finally{Pop-Location}
 & $Go test -p 2 -buildvcs=false ./...;if($LASTEXITCODE){throw 'Go tests failed'}
 & $Go vet ./...;if($LASTEXITCODE){throw 'Go vet failed'}
 & $Go run github.com/akavel/rsrc@v0.10.2 -arch amd64 -ico (Join-Path $root 'docs/assets/logo.ico') -o (Join-Path $root ('cmd/my-'+$id+'/rsrc_windows_amd64.syso'))
 if($LASTEXITCODE){throw 'Native icon export failed'}
 if($Release -and (& git -c $safe status --porcelain)){throw 'Generated source/resource differs from its commit; review and commit it'}
 & $Go build -buildvcs=false -trimpath -ldflags "-s -w -H=windowsgui -X main.version=$Version" -o (Join-Path $Output $exe) ('./cmd/my-'+$id)
 if($LASTEXITCODE){throw 'Windows x64 build failed'}
 $stage=Join-Path $Output ('package-'+[guid]::NewGuid().ToString('N'))
 New-Item -ItemType Directory -Path $stage|Out-Null
 Copy-Item -LiteralPath (Join-Path $Output $exe) -Destination $stage
 foreach($name in @('LICENSE','NOTICE','README.md','SECURITY.md','compatibility.json')){Copy-Item -LiteralPath (Join-Path $root $name) -Destination $stage}
 Copy-Item -LiteralPath (Join-Path $root 'docs') -Destination $stage -Recurse
 $contract.version=$Version
 [IO.File]::WriteAllText((Join-Path $stage 'compatibility.json'),($contract|ConvertTo-Json -Depth 10),[Text.UTF8Encoding]::new($false))
 $notice=[Text.StringBuilder]::new();$components=[Collections.Generic.List[object]]::new()
 [void]$notice.AppendLine(($contract.name+' third-party attribution. Complete license texts follow.')).AppendLine()
 $modules=& $Go list -buildvcs=false -deps -f '{{if .Module}}{{.Module.Path}}|{{.Module.Version}}|{{.Module.Dir}}{{end}}' ('./cmd/my-'+$id) | Where-Object {$_}|Sort-Object -Unique
 if($LASTEXITCODE){throw 'Go dependency inventory failed'}
 foreach($line in $modules){$parts=$line.Split('|');if($parts[0] -eq ('github.com/'+$repo)){continue};$components.Add(@{ecosystem='go';name=$parts[0];version=$parts[1]});$files=@(Get-ChildItem -LiteralPath $parts[2] -File|Where-Object Name -Match '^(LICENSE|LICENCE|COPYING|NOTICE)');if(-not $files.Count){throw ('Missing Go license: '+$parts[0])};[void]$notice.AppendLine(('===== '+$parts[0]+' '+$parts[1]+' ====='));foreach($file in $files){[void]$notice.AppendLine([IO.File]::ReadAllText($file.FullName))}}
 $goroot=(& $Go env GOROOT).Trim();[void]$notice.AppendLine('===== Go runtime =====').AppendLine([IO.File]::ReadAllText((Join-Path $goroot 'LICENSE')))
 $lock=Get-Content -LiteralPath (Join-Path $root 'web/package-lock.json') -Raw|ConvertFrom-Json
 foreach($item in $lock.packages.PSObject.Properties){if(-not $item.Name){continue};$pkg=$item.Value;$name=$item.Name -replace '^.*node_modules/','';$components.Add(@{ecosystem='npm';name=$name;version=$pkg.version;dev=[bool]$pkg.dev;optional=[bool]$pkg.optional;license=$pkg.license;integrity=$pkg.integrity});$dir=Join-Path $root ('web/'+$item.Name);if($pkg.dev -and $name -ne '@codedia/parser'){continue};if(-not(Test-Path -LiteralPath $dir)){continue};$files=@(Get-ChildItem -LiteralPath $dir -File|Where-Object Name -Match '^(LICENSE|LICENCE|COPYING|NOTICE)');if($name -eq 'pdfjs-dist'){$files+=@(Get-ChildItem -LiteralPath $dir -File -Recurse|Where-Object {$_.FullName -notmatch '[\\/]node_modules[\\/].*[\\/]node_modules[\\/]' -and $_.Name -match '^(LICENSE|LICENCE|COPYING|NOTICE)'})};if(-not $files.Count){throw ('Missing browser dependency license: '+$name)};[void]$notice.AppendLine(('===== '+$name+' '+$pkg.version+' ====='));foreach($file in ($files|Sort-Object FullName -Unique)){[void]$notice.AppendLine([IO.File]::ReadAllText($file.FullName))}}
 [IO.File]::WriteAllText((Join-Path $stage 'THIRD-PARTY-LICENSES.txt'),$notice.ToString(),[Text.UTF8Encoding]::new($false))
 [IO.File]::WriteAllText((Join-Path $stage 'dependencies.json'),(@{schema=1;application=$contract.name;version=$Version;components=$components}|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($false))
 $foundation=(& $Go list -m github.com/oneworkspacestudio/one-workspace).Trim()
 [IO.File]::WriteAllText((Join-Path $stage 'build-info.json'),(@{repository=$repo;version=$Version;sourceCommit=$source;foundation=$foundation;go=(& $Go version).Trim();architecture='windows-amd64';githubActionsUsed=$false;signing='unsigned preview';acceptance='docs/ACCEPTANCE.md'}|ConvertTo-Json -Depth 8),[Text.UTF8Encoding]::new($false))
 Add-Type -AssemblyName System.IO.Compression.FileSystem
 $archive=Join-Path $Output ('my-'+$id+'-'+$Version+'-windows-amd64.zip')
 if(Test-Path -LiteralPath $archive){throw 'Use a fresh output folder; refusing to replace a packaged archive'}
 [IO.Compression.ZipFile]::CreateFromDirectory($stage,$archive,[IO.Compression.CompressionLevel]::Optimal,$false)
 $sums=foreach($file in @((Join-Path $Output $exe),$archive)){(Get-FileHash -LiteralPath $file -Algorithm SHA256).Hash.ToLowerInvariant()+'  '+[IO.Path]::GetFileName($file)}
 [IO.File]::WriteAllLines((Join-Path $Output 'SHA256SUMS.txt'),$sums,[Text.UTF8Encoding]::new($false))
 [IO.File]::WriteAllText((Join-Path $Output 'build-receipt.json'),(@{repository=$repo;version=$Version;sourceCommit=$source;stage=$stage;archive=$archive;sha256=(Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()}|ConvertTo-Json),[Text.UTF8Encoding]::new($false))
 Write-Output ('Packaged '+$archive)
}finally{Pop-Location}
