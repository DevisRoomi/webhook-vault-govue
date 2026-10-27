@echo off
echo ========================================================
echo  WEBHOOK VAULT // INGRESS LEDGER & RETRY DISPATCHER
echo  Raw Monospaced Wireframe Fullstack Golang + Vue 3
echo ========================================================

set GOROOT=C:\Users\Lenovo\scoop\apps\go\current\go
set PATH=%GOROOT%\bin;C:\Users\Lenovo\.config\herd\bin;%PATH%

echo [1/3] Memeriksa binary backend...
if not exist "backend\vault-server.exe" (
    echo Compiling backend Go binary...
    cd backend
    go build -o vault-server.exe main.go
    cd ..
)

echo [2/3] Menjalankan Webhook Vault API pada http://127.0.0.1:8086...
start "" "backend\vault-server.exe"

echo [3/3] Membuka antarmuka Laravel Herd...
start "" "http://webhook-vault-govue.test"

echo ========================================================
echo  OPERATIONAL! 
echo  Herd Site: http://webhook-vault-govue.test
echo  Direct API: http://127.0.0.1:8086
echo  Operator demo: dispatcher / vault2026
echo ========================================================
pause
