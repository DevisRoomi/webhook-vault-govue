<?php
// Webhook Vault Laravel Herd Gateway & Reverse Proxy
$uri = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);

if (str_starts_with($uri, '/api/')) {
    $targetUrl = 'http://127.0.0.1:8086' . $_SERVER['REQUEST_URI'];
    
    $ch = curl_init($targetUrl);
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $_SERVER['REQUEST_METHOD']);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HEADER, true);
    curl_setopt($ch, CURLOPT_TIMEOUT, 15);
    
    // Forward Request Headers
    $headers = [];
    foreach (getallheaders() as $key => $value) {
        if (!in_array(strtolower($key), ['host', 'content-length'])) {
            $headers[] = "$key: $value";
        }
    }
    curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
    
    // Forward Body
    $body = file_get_contents('php://input');
    if (!empty($body)) {
        curl_setopt($ch, CURLOPT_POSTFIELDS, $body);
    }
    
    $response = curl_exec($ch);
    if ($response === false) {
        http_response_code(502);
        header('Content-Type: application/json');
        echo json_encode([
            'success' => false,
            'error' => 'Backend gateway connection refused. Ensure vault-server.exe is listening on port 8086.'
        ]);
        exit;
    }
    
    $headerSize = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    $rawHeaders = substr($response, 0, $headerSize);
    $responseBody = substr($response, $headerSize);
    curl_close($ch);
    
    http_response_code($httpCode);
    foreach (explode("\r\n", $rawHeaders) as $headerLine) {
        if (!empty($headerLine) && !str_starts_with(strtolower($headerLine), 'transfer-encoding:')) {
            header($headerLine);
        }
    }
    echo $responseBody;
    exit;
}

// Serve static frontend assets
$filePath = __DIR__ . $uri;
if ($uri !== '/' && file_exists($filePath) && !is_dir($filePath)) {
    $ext = pathinfo($filePath, PATHINFO_EXTENSION);
    $mimes = [
        'js' => 'application/javascript',
        'css' => 'text/css',
        'json' => 'application/json',
        'png' => 'image/png',
        'jpg' => 'image/jpeg',
        'svg' => 'image/svg+xml',
        'ico' => 'image/x-icon',
        'woff2' => 'font/woff2'
    ];
    if (isset($mimes[$ext])) {
        header('Content-Type: ' . $mimes[$ext]);
    }
    readfile($filePath);
    exit;
}

// Fallback to SPA index.html
if (file_exists(__DIR__ . '/index.html')) {
    header('Content-Type: text/html; charset=utf-8');
    readfile(__DIR__ . '/index.html');
    exit;
}

echo "Frontend bundle not compiled yet. Run 'bun run build' inside frontend/.";
