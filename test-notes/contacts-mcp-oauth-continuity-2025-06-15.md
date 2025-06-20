---
title: contacts-mcp-oauth-continuity-2025-06-15
type: note
permalink: basic-memory/contacts-mcp-oauth-continuity-2025-06-15
tags:
- '["contacts-mcp"'
- '"oauth"'
- '"continuity"'
- '"phase-2"'
- '"session-handoff"]'
---

# Contacts MCP OAuth - Session Continuity 2025-06-15

## Current Status
OAuth flow is implemented and working via curl tests, but Claude.ai integration showing "Invalid redirect_uri for this client" error.

## Progress Made
1. ✅ Implemented Dynamic Client Registration (RFC 7591)
2. ✅ Fixed stateless client validation 
3. ✅ OAuth endpoints working (authorize, token, callback)
4. ✅ Successful manual tests with curl
5. ❌ Claude.ai integration failing on redirect_uri validation

## Current Issue
Error: "Invalid redirect_uri for this client"

### Analysis
The redirect URI validation is too strict. Currently allows:
- `https://claude.ai/oauth/callback`
- `https://claude.ai/auth/callback`

Claude.ai might be using:
- Different subdomain (www.claude.ai, app.claude.ai)
- Different path structure
- Additional query parameters

### Next Steps
1. Add logging to capture the exact redirect_uri Claude.ai sends
2. Update validation logic to be more flexible
3. Test with actual Claude.ai integration

## Key Files
- `/src/dynamic-registration.ts` - Client registration and validation
- `/src/worker.ts` - Main request routing
- `/src/oauth.ts` - OAuth flow implementation

## Deployment
- Live URL: https://contact-mcp-ts.mike-7b5.workers.dev
- GitHub: pdxmph/contact-mcp-ts
- Issue #14 tracking this work

## Debug Commands
```bash
# Test registration
curl -X POST https://contact-mcp-ts.mike-7b5.workers.dev/register \
  -H "Content-Type: application/json" \
  -d '{"client_name": "Claude.ai", "redirect_uris": ["https://claude.ai/oauth/callback"]}'

# Test authorization
curl "https://contact-mcp-ts.mike-7b5.workers.dev/authorize?client_id=mcp_0123456789abcdef&redirect_uri=https://claude.ai/oauth/callback&response_type=code&state=test"

# Check logs
npx wrangler tail
```

## Solution Found and Deployed! ✅

### Issue Resolved
The "Invalid redirect_uri for this client" error was caused by Claude.ai using a different redirect URI path than we were validating:
- **Expected**: `/oauth/callback` or `/auth/callback`
- **Actual**: `/api/mcp/auth_callback`

### Fix Applied
Updated `isValidRedirectUri()` in `dynamic-registration.ts` to accept:
- `/oauth/*` paths
- `/auth/*` paths  
- `/api/mcp/*` paths (NEW)

### Current Status
- ✅ OAuth flow should now work with Claude.ai
- ✅ All endpoints properly configured
- ✅ Dynamic Client Registration working
- ✅ Redirect URI validation fixed

### Ready to Test
The integration should now work when adding the server URL in Claude.ai settings.

### Observations
- [development] Claude.ai uses PKCE flow with code_challenge and code_challenge_method parameters
- [implementation] Claude.ai's redirect URI follows a different pattern than typical OAuth providers
- [solution] Simple fix - just needed to capture actual request and update validation

## OAuth Implementation Complete! 🎉

### Two Issues Fixed:
1. **Redirect URI validation** - Claude.ai uses `/api/mcp/auth_callback`
2. **Token validation** - SSE endpoint expected 32-char tokens but we generate 64-char

### Final Working Configuration:
- OAuth flow completes successfully
- SSE connection established with Bearer token
- All endpoints properly configured and accepting connections

### Key Learnings:
- [debugging] Always monitor logs to see actual vs expected values
- [oauth] Claude.ai uses PKCE flow with code_challenge/code_verifier
- [implementation] Token format consistency is critical across endpoints

## OAuth Flow Complete but Connection Not Established

### Current Status - Very Close!
- ✅ OAuth flow completes successfully
- ✅ All validation issues fixed
- ❌ Claude.ai shows "connected" briefly then "not connected"
- ❌ No SSE connection attempts after OAuth

### Key Issue
Claude.ai completes the OAuth flow but never attempts to connect to the SSE endpoint. The token exchange works, but something prevents the actual MCP connection.

### Possible Causes
1. Missing token validation/introspection endpoint
2. PKCE validation not implemented
3. Session binding between OAuth and SSE missing
4. Claude.ai expecting different token format or metadata
5. UI activation step needed after OAuth

### Research Findings
- Newer MCP spec uses "Streamable HTTP" instead of SSE
- Some implementations use 15-minute token expiration
- Cloudflare provides OAuth wrappers for MCP servers
- Official implementations from Atlassian, Plaid show similar patterns

### Next Investigation Steps
- Check if Claude.ai UI has activation button after OAuth
- Implement token introspection endpoint
- Add proper PKCE validation
- Set explicit token expiration times
