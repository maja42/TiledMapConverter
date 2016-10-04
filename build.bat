@echo off

echo Performing minimal build
go build github.com/maja42/TiledMapConverter 				|| goto :error

rem Command which can be used for faster iteration in conjuction with the map pipeline. The following command can be used to deploy the exe to its final destination.
rem xcopy /y TiledMapConverter.exe "path to deploy to" || goto :error

move /y TiledMapConverter.exe export\TiledMapConverter.exe 	|| goto :error
xcopy /s/e/h/k/y resources export\resources\ 				|| goto :error

goto :EOF
:error
echo Failed with error #%errorlevel%.
pause
exit /b %errorlevel%
