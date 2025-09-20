# Flow

## Authentication Flow

```plaintext
[User Login]
    |
    v
[FE sends credentials] ---> [BE validates, responds with:]
                                 |-- access_token (JSON)
                                 |-- refresh_token (HttpOnly cookie)
    |
    v
[FE stores access_token in memory]
    |
    v
[FE makes API requests with access_token in Authorization header]
    |
    v
[access_token expires]
    |
    v
[FE gets 401, sends POST /v1/refresh with credentials: 'include']
    |
    v
[BE validates refresh_token from cookie, issues new tokens]
    |
    v
[FE updates access_token in memory, continues]
    |
    v
[User logs out]
    |
    v
[FE sends /v1/logout, BE clears refresh_token cookie, FE clears access_token]
```

## Frontend Interaction

Pseudo-code for frontend interaction with the backend API:

```js
let accessToken = null;

// Login
async function login(email, password) {
  const res = await fetch('/v1/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include', // so refresh_token cookie is set
  });
  const data = await res.json();
  accessToken = data.access_token;
}

// Authenticated API call
async function fetchUserProfile() {
  const res = await fetch('/v1/users/me', {
    headers: { Authorization: `Bearer ${accessToken}` },
    credentials: 'include',
  });
  if (res.status === 401) {
    // Try to refresh token
    await refreshToken();
    // Retry original request
    return fetchUserProfile();
  }
  return res.json();
}

// Refresh token
async function refreshToken() {
  const res = await fetch('/v1/refresh', {
    method: 'POST',
    credentials: 'include', // sends refresh_token cookie
  });
  if (res.ok) {
    const data = await res.json();
    accessToken = data.access_token;
  } else {
    // Redirect to login
    accessToken = null;
    window.location.href = '/login';
  }
}

// Logout
async function logout() {
  await fetch('/v1/logout', {
    method: 'POST',
    credentials: 'include',
  });
  accessToken = null;
  window.location.href = '/login';
}
```

or

```js
let accessToken = null;

// On login
async function login(email, password) {
  // ...same as before...
  accessToken = data.access_token;
}

// On API call
async function apiCall(url, options = {}) {
  options.headers = options.headers || {};
  options.headers.Authorization = `Bearer ${accessToken}`;
  let res = await fetch(url, options);

  if (res.status === 401) {
    // Try to refresh
    const refreshed = await refreshToken();
    if (refreshed) {
      options.headers.Authorization = `Bearer ${accessToken}`;
      res = await fetch(url, options); // retry
    } else {
      // Redirect to login
      window.location.href = '/login';
      return;
    }
  }
  return res.json();
}

// On refresh (web: cookie, mobile: header/body)
async function refreshToken() {
  let res;
  if (isMobileApp) {
    // Send refresh token from secure storage
    const refreshToken = await getRefreshTokenFromSecureStorage();
    res = await fetch('/v1/refresh', {
      method: 'POST',
      headers: { 'X-Refresh-Token': refreshToken },
    });
  } else {
    // Web: cookie is sent automatically
    res = await fetch('/v1/refresh', {
      method: 'POST',
      credentials: 'include',
    });
  }
  if (res.ok) {
    const data = await res.json();
    accessToken = data.access_token;
    return true;
  }
  return false;
}
```

In web application situations, the `access_token` is typically stored in memory (e.g., a React state or context) to minimize exposure to XSS attacks. The `refresh_token` is stored in an HttpOnly cookie, which is not accessible via JavaScript, providing protection against XSS. The backend validates the `refresh_token` on each refresh request to ensure session integrity.

Key points:

- The FE never directly accesses the refresh token; it’s managed by the browser as a cookie.
- The access token is kept in memory and used for API calls.
- When the access token expires, the FE uses the refresh token (via cookie) to get a new access token.
- Logout clears both tokens.

## Refactor RefreshTokenHandler() to support refresh token not only from cookie but also from request body or even X-Refresh-Token header

```go
func RefreshTokenHandler(c *fiber.Ctx) error {
    var refreshTokenString string

    // 1. Try custom header
    refreshTokenString = c.Get("X-Refresh-Token")

    // 2. Try request body if not found in header
    if refreshTokenString == "" {
        var body struct {
            RefreshToken string `json:"refresh_token"`
        }
        if err := c.BodyParser(&body); err == nil && body.RefreshToken != "" {
            refreshTokenString = body.RefreshToken
        }
    }

    // 3. Fallback to cookie
    if refreshTokenString == "" {
        refreshTokenString = c.Cookies("refresh_token")
    }

    if refreshTokenString == "" {
        return response.SendErrorResponse(c, fiber.StatusUnauthorized, "no_refresh_token")
    }

    // ...existing logic...
}
```

```js
<AuthProvider>
  <Routes>
    <Route path="/login" element={<LoginPage />} />
    <Route path="/dashboard" element={
      <RequireAuth>
        <DashboardPage />
      </RequireAuth>
    } />
    {/* ...other routes... */}
  </Routes>
</AuthProvider>
```