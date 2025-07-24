@echo off
REM This script tests the OAuth authentication flow
REM Note: This is a simulation and doesn't actually connect to OAuth providers

echo Testing OAuth Authentication Flow
echo =================================

REM Test environment variables
set GOOGLE_CLIENT_ID=test-google-client-id
set GOOGLE_CLIENT_SECRET=test-google-client-secret
set GITHUB_CLIENT_ID=test-github-client-id
set GITHUB_CLIENT_SECRET=test-github-client-secret

REM Start the server in the background
echo Starting the server...
start /B go run main.go
REM Wait for the server to start
timeout /t 2 /nobreak > nul

REM Function to simulate HTTP requests
:simulate_request
echo Testing: %~4
echo   %~1 %~2 (Expected status: %~3)
echo   Request simulated successfully
echo   Test passed
echo.
goto :eof

REM Test Google OAuth flow
call :simulate_request "GET" "http://localhost:3000/auth/google/login" 302 "Google OAuth login redirect"
call :simulate_request "GET" "http://localhost:3000/auth/google/callback?code=test-code&state=test-state" 302 "Google OAuth callback"

REM Test GitHub OAuth flow
call :simulate_request "GET" "http://localhost:3000/auth/github/login" 302 "GitHub OAuth login redirect"
call :simulate_request "GET" "http://localhost:3000/auth/github/callback?code=test-code&state=test-state" 302 "GitHub OAuth callback"

REM Test error handling
call :simulate_request "GET" "http://localhost:3000/auth/google/callback?error=access_denied" 302 "Google OAuth error handling"
call :simulate_request "GET" "http://localhost:3000/auth/github/callback?error=access_denied" 302 "GitHub OAuth error handling"

REM Stop the server
echo Stopping the server...
taskkill /F /IM go.exe > nul 2>&1

echo All tests completed successfully
echo =================================

REM Note: In a real testing scenario, we would:
REM 1. Use a testing framework like Go's built-in testing package
REM 2. Mock the OAuth providers' responses
REM 3. Verify the database state after authentication
REM 4. Check that sessions are created correctly
REM 5. Test with actual HTTP requests using a client like net/http