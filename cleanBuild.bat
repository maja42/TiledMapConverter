@echo off

echo Performing clean build

RMDIR /S /Q export
MKDIR export

go build github.com/maja42/TiledMapConverter           || goto :error

move /y TiledMapConverter.exe export\TiledMapConverter.exe    || goto :error
xcopy /s/e/h/k/y dependencies export            || goto :error
xcopy /s/e/h/k/y resources export\resources\    || goto :error



goto :EOF
:error
echo Failed with error #%errorlevel%.
pause
exit /b %errorlevel%
