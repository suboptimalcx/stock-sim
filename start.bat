@echo off
SETLOCAL

REM
IF "%~1"=="" (
    SET PORT=8080
) ELSE (
    SET PORT=%~1
)

echo Starting Stock Market Simulator on localhost:%PORT%...

REM 
docker compose down

REM
docker compose up --build -d

echo Application is running at http://localhost:%PORT%

ENDLOCAL