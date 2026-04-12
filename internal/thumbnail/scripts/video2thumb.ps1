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

$ffmpegparams = "%%FFMPEG_PARAMS%% -ss 00:00:35 -i $($Source) -frames:v 1 $($Destination).png"
Start-Process -FilePath %%FFMPEG%% -ArgumentList $ffmpegparams -NoNewWindow -Wait

$convertParams = "%%CONVERT_PARAMS%% $($Destination).png -resize $($Width)x$($Height) -background $($Background) -gravity Center -extent $($Width)x$($Height) $($Destination)"
Start-Process -FilePath %%CONVERT%% -ArgumentList $convertparams -NoNewWindow -Wait

Remove-Item -Path "$($Destination).png"
