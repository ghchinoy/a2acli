# Authentication & Security Schemes Guide

A2A agents declare their authentication requirements in the `AgentCard` and enforce them at request time.

---

## 1. Security Schemes in the Agent Card

Support one or more standard schemes under `AgentCard.securitySchemes`:

### HTTP Bearer JWT
```json
"securitySchemes": {
  "bearerAuth": {
    "type": "httpAuthSecurityScheme",
    "scheme": "bearer",
    "bearerFormat": "JWT"
  }
}
```

### API Key
```json
"securitySchemes": {
  "apiKeyAuth": {
    "type": "apiKeySecurityScheme",
    "name": "X-API-Key",
    "location": "header"
  }
}
```

### OAuth 2.1 (Authorization Code + PKCE)
```json
"securitySchemes": {
  "oauth2": {
    "type": "oauth2SecurityScheme",
    "flows": {
      "authorizationCode": {
        "authorizationUrl": "https://auth.example.com/oauth/authorize",
        "tokenUrl": "https://auth.example.com/oauth/token",
        "scopes": { "agent:invoke": "Invoke agent skills" },
        "pkceRequired": true
      }
    }
  }
}
```

---

## 2. Server Authentication Responsibilities (§7.4)

1. **Extract & Validate Credentials:** On every request, extract token/key from HTTP header, gRPC metadata, or query parameter.
2. **Reject Unauthenticated Calls:** Return HTTP 401 (`ErrUnauthenticated` / gRPC `UNAUTHENTICATED`) if credentials are missing or invalid.
3. **Inject Authenticated User Context:** Attach `User` object to request context (`a2asrv.CallContext.User` in Go, `ServerCallContext.user` in Python).
4. **Skill-Level Authorization:** Check if caller has permission for the requested skill ID.

---

## 3. In-Task Authorization (§7.6)

When an agent needs third-party user credentials or OAuth authorization during task processing:

1. Transition task state to `TASK_STATE_AUTH_REQUIRED`.
2. Include a `TaskStatus` message explaining the required authorization URL or action.
3. Accept out-of-band credential receipt or follow-up message.
4. Resume task processing (`TASK_STATE_WORKING` → `TASK_STATE_COMPLETED`).
