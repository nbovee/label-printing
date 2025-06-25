@echo off
echo Building Label Printer (GUI mode)...
go build -ldflags="-H windowsgui" -o label-printer.exe main.go
if %ERRORLEVEL% EQU 0 (
    echo Build successful! Run label-printer.exe to start the application.
    echo Note: No console window will appear when running from Explorer.
) else (
    echo Build failed!
)
pause 