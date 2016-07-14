@echo off

echo Fetching go dependencies...

go get "github.com/op/go-logging" || goto :error

echo All dependencies have been retrieved
pause

goto :EOF
:error
echo Failed with error #%errorlevel%.
pause

