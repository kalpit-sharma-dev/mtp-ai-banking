@echo off
REM Script to check which ports are in use

echo ==========================================
echo Checking Port Usage
echo ==========================================
echo.

echo Checking ports used by AI Banking Platform...
echo.

for %%p in (8080 8081 8001 8002 8003 7000 9000) do (
    echo Port %%p:
    netstat -ano | findstr :%%p
    if errorlevel 1 (
        echo   [FREE]
    ) else (
        echo   [IN USE]
    )
    echo.
)

pause

