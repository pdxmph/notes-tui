---
title: contact-mcp-oauth-debug-continuity-2025-06-15
type: note
permalink: basic-memory/contact-mcp-oauth-debug-continuity-2025-06-15
---

# Contact MCP OAuth Debug Session Continuity - 2025-06-15

## Context
Working to fix tools not appearing in Claude.ai for pdxmph/contact-mcp-ts OAuth-enabled MCP server.

## Current Status
- ✅ OAuth flow working perfectly
- ✅ Protocol version matching (2024-11-05)
- ✅ Basic initialize handshake working
- ❌ Tools still not appearing in Claude.ai

## Key Discoveries

### 1. Protocol Version Mismatch (FIXED)
- Claude.ai requests `2024-11-05`, we were responding with `2025-03-26`
- Now echoing back the client's requested version

### 2. Tools Discovery Pattern
**The Problem**: Claude.ai never makes a `tools/list` request, despite MCP spec saying that's how tools are discovered.

**What We Tried**:
- `"tools": {}` - Handshake works, no tools appear
- `"tools": [array]` - Breaks handshake completely
- `"tools": { "listChanged": true }` - Should be correct per spec, but still no tools

### 3. Communication Pattern
Claude.ai's actual flow:
1. POST / with initialize
2. POST / with notifications/initialized
3. GET / for SSE (hangs/times out)
4. POST /messages?sessionId=... for subsequent requests

## Technical Details
- GitHub repo: pdxmph/contact-mcp-ts
- Local: /Users/mph/code/contact-mcp-ts
- Live: https://contact-mcp-ts.mike-7b5.workers.dev
- Main issue: #14

## Next Session Focus
**Stop guessing, start studying working examples**

Need to find OAuth-enabled MCP servers that successfully show tools in Claude.ai and understand their implementation. The spec isn't helping us understand what Claude.ai actually expects.

## Key Questions for Next Session
1. Are there any public OAuth-enabled MCP servers that work with Claude.ai?
2. What's different about their tool discovery mechanism?
3. Is the SSE connection critical for tool discovery?
4. Does the /messages endpoint play a role in tool discovery?