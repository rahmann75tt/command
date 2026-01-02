$searchPath = "{PATH}"
$searchPathFull = [System.IO.Path]::GetFullPath($searchPath)
$depthStr = "{DEPTH}"
$cmd = if ($depthStr -eq "unlimited") {
	Get-ChildItem -Force -Path $searchPath -Recurse
} else {
	Get-ChildItem -Force -Path $searchPath -Depth ([int]$depthStr) -Recurse
}
$cmd | ForEach-Object {
	$fullPath = $_.FullName
	# Make path relative to search path root
	if ($fullPath.StartsWith($searchPathFull)) {
		$relPart = $fullPath.Substring($searchPathFull.Length).TrimStart('\')
		if ($relPart) {
			$outputPath = $searchPath.TrimEnd('\') + "\" + $relPart
		} else {
			$outputPath = $searchPath
		}
	} else {
		$outputPath = $fullPath
	}
	$modTime = $_.LastWriteTime.ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
	$US = [char]0x1F  # Unit Separator (field separator)
	$RS = [char]0x1E  # Record Separator
	if ($_.PSIsContainer) {
		Write-Host -NoNewline "$($_.Name)${US}DIR${US}$modTime${US}$outputPath${RS}"
	} else {
		Write-Host -NoNewline "$($_.Name)${US}$($_.Length)${US}$modTime${US}$outputPath${RS}"
	}
}
