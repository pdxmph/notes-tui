---
title: OAuth Debug - Claude.ai Integration Issue
type: note
permalink: basic-memory/oauth-debug-claude-ai-integration-issue
---

# OAuth Debug - Claude.ai Integration Issue

## Problem
Claude.ai shows error: "There was an error connecting to Contacts server. Please check your server URL and make sure your server handles auth correctly."

## Current OAuth Metadata
Our server returns:
```json
{
    "issuer": "https://contact-mcp-ts.mike-7b5.workers.dev",
    "authorization_endpoint": "https://contact-mcp-ts.mike-7b5.workers.dev/authorize",
    "token_endpoint": "https://contact-mcp-ts.mike-7b5.workers.dev/token",
    "registration_endpoint": "https://contact-mcp-ts.mike-7b5.workers.dev/register",
    "token_endpoint_auth_methods_supported": ["client_secret_post", "none"],
    "response_types_supported": ["code"],
    "response_modes_supported": ["query"],
    "grant_types_supported": ["authorization_code"],
    "code_challenge_methods_supported": ["S256"],
    "scopes_supported": ["mcp"],
    "service_documentation": "https://github.com/pdxmph/contact-mcp-ts",
    "mcp_transports_supported": ["sse"],
    "mcp_endpoint": "https://contact-mcp-ts.mike-7b5.workers.dev/mcp/oauth",
    "mcp_version": "2024-11-05"
}
```

## Hypothesis
Claude.ai might be:
1. Expecting a specific client_id in the metadata
2. Looking for different OAuth fields
3. Unable to handle the dynamic client registration rejection
4. Expecting the callback URL to be pre-configured

## Potential Solutions
1. Add client metadata to OAuth discovery
2. Support dynamic client registration properly
3. Add more specific error responses
4. Check if Claude.ai needs allowlisted redirect URIs