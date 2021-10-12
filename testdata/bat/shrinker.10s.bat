@echo off

echo "grow-shrink"
echo "---"
set /a num=%random% %%10 %+1
REM echo "tiems "
REM echo %num%
for /l %%x in (1, 1, %num%) do echo "ohyeah up to %num%"
