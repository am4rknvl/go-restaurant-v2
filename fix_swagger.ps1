# PowerShell script to fix Swagger documentation issues

Write-Host "Fixing Swagger documentation issues..." -ForegroundColor Green

# Get all Go files in handlers directory
$files = Get-ChildItem -Path ".\internal\handlers\*.go" -Recurse

$totalFixed = 0

foreach ($file in $files) {
    $content = Get-Content $file.FullName -Raw
    $originalContent = $content
    
    # Fix double @ to single @
    $content = $content -replace '// @Failure', '// @Failure'
    $content = $content -replace '// @Success', '// @Success'
    $content = $content -replace '// @Param', '// @Param'
    
    # Fix ErrorRespons to ErrorResponse
    $content = $content -replace 'models\.ErrorRespons', 'models.ErrorResponse'
    
    # Only write if content changed
    if ($content -ne $originalContent) {
        Set-Content -Path $file.FullName -Value $content -NoNewline
        Write-Host "Fixed: $($file.Name)" -ForegroundColor Yellow
        $totalFixed++
    }
}

Write-Host "`nTotal files fixed: $totalFixed" -ForegroundColor Green
Write-Host "Now run: swag init" -ForegroundColor Cyan
