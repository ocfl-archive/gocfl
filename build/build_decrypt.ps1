$archs = "amd64","arm64"
$oss = "windows","linux","darwin"

foreach ($os in $oss) {
    $env:GOOS=$os
    foreach ($arch in $archs) {
        $env:GOARCH=$arch
        Write-Output "$os/$arch"
        if ($os -eq "windows") {
            Start-Process -FilePath "go.exe" -ArgumentList "build -o ./decrypt_$($os)_$($arch).exe ../decrypt" -Wait
        } else {
            Start-Process -FilePath "go.exe" -ArgumentList "build -o ./decrypt_$($os)_$($arch) ../decrypt" -Wait
        }
    }
}