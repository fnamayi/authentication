# OAuth Authentication

This document provides information about the OAuth authentication implementation in the forum application.

## Overview

The forum application now supports OAuth authentication with Google and GitHub. This allows users to sign in using their existing accounts on these platforms, providing a more convenient and secure authentication experience.

## Supported OAuth Providers

- **Google**: Users can sign in with their Google accounts.
- **GitHub**: Users can sign in with their GitHub accounts.

## How It Works

1. **Initiation**: When a user clicks on an OAuth button on the login page, they are redirected to the provider's authorization page.
2. **Authorization**: The user authorizes the application to access their basic profile information.
3. **Callback**: After authorization, the provider redirects back to our application with an authorization code.
4. **Token Exchange**: The application exchanges the authorization code for an access token.
5. **User Information**: The application uses the access token to fetch the user's information from the provider.
6. **Account Creation/Login**: If the user doesn't exist in our database, a new account is created. Otherwise, the user is logged in to their existing account.

## Implementation Details

### Database Schema

The users table has been extended with the following fields:

- `oauth_provider`: The name of the OAuth provider (e.g., "google", "github").
- `oauth_id`: The unique ID of the user from the OAuth provider.

### OAuth Flow

1. **Login Initiation**:
   - `/auth/google/login`: Initiates the Google OAuth flow.
   - `/auth/github/login`: Initiates the GitHub OAuth flow.

2. **Callback Handling**:
   - `/auth/google/callback`: Handles the callback from Google.
   - `/auth/github/callback`: Handles the callback from GitHub.

### Security Considerations

- **CSRF Protection**: The OAuth flow includes state parameter validation to prevent cross-site request forgery attacks.
- **Secure Cookies**: Session cookies are marked as HTTP-only to prevent client-side access.
- **Minimal Scope**: Only the minimal required scopes are requested from the OAuth providers.

## Configuration

To use OAuth authentication, you need to set up the following environment variables:

- **Google OAuth**:
  - `GOOGLE_CLIENT_ID`: Your Google OAuth client ID.
  - `GOOGLE_CLIENT_SECRET`: Your Google OAuth client secret.

- **GitHub OAuth**:
  - `GITHUB_CLIENT_ID`: Your GitHub OAuth client ID.
  - `GITHUB_CLIENT_SECRET`: Your GitHub OAuth client secret.

## Setting Up OAuth Credentials

### Google OAuth Setup

1. Go to the [Google Cloud Console](https://console.cloud.google.com/).
2. Create a new project or select an existing one.
3. Navigate to "APIs & Services" > "Credentials".
4. Click "Create Credentials" > "OAuth client ID".
5. Select "Web application" as the application type.
6. Add "http://localhost:3000/auth/google/callback" as an authorized redirect URI.
7. Copy the client ID and client secret.

### GitHub OAuth Setup

1. Go to your [GitHub Settings](https://github.com/settings/developers).
2. Click on "OAuth Apps" > "New OAuth App".
3. Fill in the application details:
   - Application name: Forum
   - Homepage URL: http://localhost:3000
   - Authorization callback URL: http://localhost:3000/auth/github/callback
4. Register the application and copy the client ID and client secret.

## User Experience

When a user visits the login page, they will see options to sign in with Google or GitHub in addition to the traditional username/password form. Clicking on one of these options will redirect them to the respective provider's authorization page.

After authorizing the application, they will be redirected back to the forum and automatically logged in.