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

$gsparams = "%%GHOSTSCRIPT_PARAMS%% -dNOPAUSE -dBATCH -sDEVICE=png16m -dFirstPage=1 -dLastPage=1 -sOutputFile=$($Destination).png $($Source)"
Start-Process -FilePath %%GHOSTSCRIPT%% -ArgumentList $gsparams -NoNewWindow -Wait

$convertParams = "%%CONVERT_PARAMS%% $($Destination).png -resize $($Width)x$($Height) -background $($Background) -gravity Center -extent $($Width)x$($Height) $($Destination)"
Start-Process -FilePath %%CONVERT%% -ArgumentList $convertparams -NoNewWindow -Wait

Remove-Item -Path "$($Destination).png"
