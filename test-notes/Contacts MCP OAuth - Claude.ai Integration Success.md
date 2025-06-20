---
title: Contacts MCP OAuth - Claude.ai Integration Success
type: note
permalink: basic-memory/contacts-mcp-oauth-claude-ai-integration-success
---

# Contacts MCP OAuth Implementation - Claude.ai Integration Success

## Summary
Successfully implemented Dynamic Client Registration (RFC 7591) to enable Claude.ai integration with the contacts MCP server.

## Problem Solved
Claude.ai requires Dynamic Client Registration support and doesn't allow users to manually specify client IDs. Our server was rejecting registration requests, preventing Claude.ai from connecting.

## Solution Implemented
1. **Created `dynamic-registration.ts`**: Full RFC 7591 compliant implementation
2. **In-memory client storage**: Stores registered clients with their metadata
3. **Updated OAuth flow**:
   - `/register`: Now accepts and processes registration requests
   - `/authorize`: Validates registered clients and redirect URIs
   - `/token`: Validates registered clients, supports public clients
4. **State encoding**: Preserves client info through the OAuth flow

## Key Code Changes
- Added support for public clients (no client_secret required)
- Implemented proper CORS headers for claude.ai origin
- Encoded client_id and redirect_uri in state parameter
- Validated redirect URIs (must be HTTPS except localhost)

## Test Results
Successfully registered a client:
```json
{
    "client_id": "mcp_b8a83514038ce493",
    "client_name": "Claude.ai",
    "redirect_uris": ["https://claude.ai/oauth/callback"],
    "token_endpoint_auth_method": "none",
    "client_id_issued_at": 1749954403
}
```

## Next Steps
- Test complete OAuth flow with Claude.ai
- Monitor for any edge cases
- Consider persistent storage for registered clients

## References
- [RFC 7591: OAuth 2.0 Dynamic Client Registration](https://datatracker.ietf.org/doc/html/rfc7591)
- [Anthropic MCP Documentation](https://support.anthropic.com/en/articles/11503834-building-custom-integrations-via-remote-mcp-servers)
- [MCP Authorization Spec](https://modelcontextprotocol.io/specification/2025-03-26/basic/authorization)