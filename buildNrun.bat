@echo off

rem Kill running application
tasklist /FI "IMAGENAME eq TiledMapConverter.exe" 2>NUL | find /I /N "TiledMapConverter.exe">NUL
if "%ERRORLEVEL%"=="0" taskkill /f /im TiledMapConverter.exe

echo "Building..."
call build.bat || goto :error

echo Executing application
echo *************************************
call export\TiledMapConverter.exe || goto :error

goto :EOF
:error
echo Failed with error #%errorlevel%.
pause
exit /b %errorlevel%