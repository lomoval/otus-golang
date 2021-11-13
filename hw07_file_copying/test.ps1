go build -o go-cp.exe

./go-cp.exe -from testdata/input.txt -to out.txt
cmp out.txt testdata/out_offset0_limit0.txt

./go-cp.exe -from testdata/input.txt -to out.txt -limit 10
cmp out.txt testdata/out_offset0_limit10.txt

./go-cp.exe -from testdata/input.txt -to out.txt -limit 1000
cmp out.txt testdata/out_offset0_limit1000.txt

./go-cp.exe -from testdata/input.txt -to out.txt -limit 10000
cmp out.txt testdata/out_offset0_limit10000.txt

./go-cp.exe -from testdata/input.txt -to out.txt -offset 100 -limit 1000
cmp out.txt testdata/out_offset100_limit1000.txt

./go-cp.exe -from testdata/input.txt -to out.txt -offset 6000 -limit 1000
cmp out.txt testdata/out_offset6000_limit1000.txt

rm go-cp.exe
rm out.txt
echo "PASS"
