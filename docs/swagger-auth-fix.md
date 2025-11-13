# Swagger Authentication Fix ✅

## Problems Fixed

### Issue 1: Invalid Authorization Type
**Symptom:** Getting `{"code": 401, "message": "invalid authorization type"}` when testing protected endpoints

**Root Cause:** The auth middleware was case-sensitive when checking "Bearer" vs "bearer"

**Solution:** Updated `internal/common/middleware/auth.go:33` to convert authorization type to lowercase before comparison

### Issue 2: Invalid Token
**Symptom:** After fixing Issue 1, still getting `{"code": 401, "message": "invalid token"}` with valid tokens

**Root Causes:**
1. **Field name mismatch**: Payload struct had `json:"expire_at"` but CreateToken used `"expired_at"`
2. **Type mismatch**: CreateToken stored Unix timestamps as `int64`, but Payload expected `time.Time`
3. **Wrong error mapping**: jwt.ErrTokenExpired was mapped to ErrInvalidToken instead of ErrExpiredToken

**Solutions:**
1. Updated `utils/token/payload.go:21` - Changed json tag from `"expire_at"` to `"expired_at"`
2. Updated `utils/token/jwt_maker.go:28` - Changed to use Payload struct directly instead of MapClaims
3. Updated `utils/token/jwt_maker.go:44` - Fixed to return ErrExpiredToken for expired tokens

---

## How to Use Authentication in Swagger UI

### Step 1: Login to Get Token

1. Go to Swagger UI: http://localhost:8080/swagger/index.html
2. Find the `POST /users/login` endpoint
3. Click "Try it out"
4. Fill in the request body:
   ```json
   {
     "email": "your@email.com",
     "password": "yourpassword"
   }
   ```
5. Click "Execute"
6. Copy the `access_token` from the response

### Step 2: Authorize in Swagger UI

**IMPORTANT:** Only enter the token itself, NOT "Bearer" prefix!

1. Click the **"Authorize"** button at the top right (🔒 icon)
2. In the popup dialog, you'll see:
   ```
   Bearer (apiKey)
   Value: [input box]
   ```
3. **Paste ONLY the token** (without "Bearer "):
   ```
   eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2Vybm...
   ```

   ❌ **DON'T** enter:
   ```
   Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```

   ✅ **DO** enter:
   ```
   eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```

4. Click "Authorize"
5. Click "Close"

### Step 3: Test Protected Endpoints

Now you can test any protected endpoint:
- GET /users/profile
- PUT /users/profile
- GET /users/sessions
- DELETE /users/sessions/{session_id}
- etc.

The authorization header will be automatically added as:
```
Authorization: Bearer YOUR_TOKEN
```

---

## Why This Works Now

### Fix 1: Case-Insensitive Authorization Type

**Before:**
```go
// auth.go (OLD)
authorizationType := fields[0]  // Could be "Bearer" or "bearer"
if authorizationType != "bearer" {  // Only matched lowercase
    return error
}
```

**After:**
```go
// auth.go (NEW)
authorizationType := strings.ToLower(fields[0])  // Convert to lowercase
if authorizationType != "bearer" {  // Always matches
    return error
}
```

Now the middleware accepts both:
- ✅ `Authorization: Bearer token`
- ✅ `Authorization: bearer token`

### Fix 2: Consistent Token Structure

**Before:**
```go
// jwt_maker.go (OLD) - CreateToken
jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "id":         payload.ID,
    "user_id":    userID,
    "username":   username,
    "role":       role,
    "issued_at":  time.Now().Unix(),     // int64
    "expired_at": time.Now().Add(duration).Unix(),  // int64
})

// payload.go (OLD)
ExpiredAt time.Time `json:"expire_at"`  // Wrong tag, wrong type
```

**After:**
```go
// jwt_maker.go (NEW) - CreateToken
jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

// payload.go (NEW)
ExpiredAt time.Time `json:"expired_at"`  // Matching tag and type
```

Now token creation and verification use the same Payload structure:
- ✅ Field names match between creation and verification
- ✅ Types match (time.Time instead of int64)
- ✅ Expired tokens return proper "token has expired" message

### Fix 3: Proper Error Handling

**Before:**
```go
// jwt_maker.go (OLD) - VerifyToken
if errors.Is(err, jwt.ErrTokenExpired) {
    return nil, ErrInvalidToken  // Wrong error!
}
```

**After:**
```go
// jwt_maker.go (NEW) - VerifyToken
if errors.Is(err, jwt.ErrTokenExpired) {
    return nil, ErrExpiredToken  // Correct error
}
```

Now expired tokens show the correct error message instead of generic "invalid token"

---

## Testing the Fix

### Option 1: Use Swagger UI (Recommended)

1. Start server: `make run`
2. Open: http://localhost:8080/swagger/index.html
3. Login and authorize (as described above)
4. Test `GET /users/profile`
5. Should return your user profile ✅

### Option 2: Use curl

```bash
# 1. Login
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# Copy access_token from response

# 2. Get profile (with Bearer - uppercase)
curl http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# 3. Get profile (with bearer - lowercase)
curl http://localhost:8080/api/v1/users/profile \
  -H "Authorization: bearer YOUR_ACCESS_TOKEN"

# Both should work now! ✅
```

### Option 3: Use httpie

```bash
# Login
http POST localhost:8080/api/v1/users/login \
  email=test@example.com \
  password=password123

# Get profile
http GET localhost:8080/api/v1/users/profile \
  "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

---

## Common Issues & Solutions

### Issue 1: Still getting 401 error

**Possible causes:**
1. Token is expired
   - Solution: Login again to get a fresh token
2. Wrong token
   - Solution: Make sure you copied the entire access_token
3. Server not restarted
   - Solution: Restart server after code changes

### Issue 2: "authorization header is not provided"

**Cause:** Forgot to authorize in Swagger UI

**Solution:**
1. Click "Authorize" button
2. Paste your token
3. Click "Authorize"

### Issue 3: "invalid authorization header format"

**Cause:** Authorization header is malformed

**Solution in Swagger UI:**
- Only paste the token, not "Bearer token"
- Swagger adds "Bearer " automatically

**Solution in curl:**
- Include "Bearer " manually:
  ```bash
  curl -H "Authorization: Bearer YOUR_TOKEN" ...
  ```

---

## Token Expiration

### Access Token
- Duration: Set in config (typically 15 minutes)
- When expired: Login again or use refresh token

### Refresh Token
- Duration: Set in config (typically 7 days)
- Use endpoint: `POST /users/refresh`
- Request body:
  ```json
  {
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }
  ```

---

## Security Best Practices

1. **Never commit tokens** to version control
2. **Use HTTPS** in production
3. **Set appropriate token expiration** times
4. **Revoke sessions** when user logs out
5. **Use refresh tokens** to minimize access token exposure

---

**Status: ✅ Fixed and Tested**

You can now authenticate successfully in Swagger UI!
