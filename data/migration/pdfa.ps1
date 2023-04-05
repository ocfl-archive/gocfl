$temp = New-TemporaryFile
$Input | Out-File $temp
$tempFolder = $temp.DirectoryName -replace '\\', '/'
$pdf = "{0}.pdf" -f $temp.FullName
Rename-Item $temp $pdf
$pdfSlash = $pdf -replace '\\', '/'
Start-Process gswin64.exe -NoNewWindow -Wait -ArgumentList "-dBATCH","-dNODISPLAY","-dNOPAUSE","-dNOSAFER","-sDEVICE=pdfwrite","-dPDFA=2","-sColorConversionStrategy=RGB","-dPDFACompatibilityPolicy=1","--permit-file-read=$($tempFolder)","-sOutputFile=-","c:/daten/go/dev/gocfl/data/migration/pdfa_def.ps","$($pdfSlash)"
#Remove-Item -Path $pdf
