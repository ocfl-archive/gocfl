Param (
    [Parameter(Mandatory=$true, ValueFromPipeline=$false)]
    [string]$Source,

    [Parameter(Mandatory=$true, ValueFromPipeline=$false)]
    [string]$Destination,

    [Parameter(Mandatory=$false, ValueFromPipeline=$false)]
    [string]$Background = "none",

    [Parameter(Mandatory=$false, ValueFromPipeline=$false)]
    [int]$Width = 256,

    [Parameter(Mandatory=$false, ValueFromPipeline=$false)]
    [int]$Height = 256
)

$ffmpegparams = "-ss 00:00:35 -i $($Source) -frames:v 1 $($Destination).png"
Start-Process -FilePath ffmpeg.exe -ArgumentList $ffmpegparams -NoNewWindow -Wait

$convertParams = "$($Destination).png -resize $($Width)x$($Height) -background $($Background) -gravity Center -extent $($Width)x$($Height) $($Destination)"
Start-Process -FilePath convert.exe -ArgumentList $convertparams -NoNewWindow -Wait

Remove-Item -Path "$($Destination).png"
