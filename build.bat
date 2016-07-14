@echo off

echo Performing minimal build
go build github.com/maja42/TiledMapConverter 				|| goto :error

move /y TiledMapConverter.exe export\TiledMapConverter.exe 	|| goto :error
xcopy /s/e/h/k/y resources export\resources\ 				|| goto :error


goto :EOF
:error
echo Failed with error #%errorlevel%.
pause
exit /b %errorlevel%
