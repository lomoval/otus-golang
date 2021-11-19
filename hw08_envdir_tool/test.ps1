# #!/usr/bin/env bash
#set -xeuo pipefail

go build -o go-envdir.exe

$env:HELLO="SHOULD_REPLACE"
$env:FOO="SHOULD_REPLACE"
$env:UNSET="SHOULD_REMOVE"
$env:ADDED="from original env"
$env:EMPTY="SHOULD_BE_EMPTY"

$result = & ./go-envdir "$($pwd.Path)/testdata/env" "powershell" "$($pwd.Path)/testdata/echo.ps1" "arg1=1" "arg2=2"
$expected=@'
HELLO is ("hello")
BAR is (bar)
FOO is (   foo
with new line)
UNSET is ()
ADDED is (from original env)
EMPTY is ()
arguments are arg1=1 arg2=2
'@

rm  go-envdir.exe

if ($result.Equals($expected))
{
    echo invalid output: $result
    exit 1
}

echo "PASS"
