@echo off

FOR /F "tokens=*" %%i IN (test_images/.TAGLIST) DO (
	FOR /F "tokens=1" %%a IN ("%%i") do (
		FOR /F "tokens=2" %%b IN ("%%i") do (
			tis add-file "./test_images/%%a" "%%b" --no-move
		)	
	)
)