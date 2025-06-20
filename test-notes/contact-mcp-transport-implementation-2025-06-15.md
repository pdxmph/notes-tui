---
title: contact-mcp-transport-implementation-2025-06-15
type: note
permalink: basic-memory/contact-mcp-transport-implementation-2025-06-15
---

# Contact MCP Transport Implementation - 2025-06-15

## Session Summary
After extensive debugging and research into why tools weren't appearing in Claude.ai, we discovered the root cause: the MCP server implementation isn't using the SDK's transport classes correctly. Instead of letting the SDK handle the protocol, we're returning "Method not implemented" for all requests.

## Key Discovery
Tools are exposed in the `initialize` response, NOT through a separate `tools/list` request. The MCP SDK's transport classes handle this automatically when used correctly. Claude.ai never makes a `tools/list` request.

## Current Status
- ✅ OAuth flow working perfectly
- ✅ MCP server created and tools registered
- ❌ Request handling not implemented (returns error for all requests)
- ❌ Not using SDK transport classes
- Created issue #21 to track implementation work

## Implementation Plan
Based on SimpleScraper and other working examples, we need to:

1. **Use StreamableHTTPServerTransport** from the MCP SDK
2. **Maintain transport instances** per session in a `transports` object
3. **Connect MCP server to transport** on initialize
4. **Let transport handle all requests** instead of manual handling

## Code Location
- File: `src/worker-mcp.ts`
- Problem area: Lines where `server.request(body)` is commented out
- Need to replace with proper transport handling

## Reference Pattern
```javascript
const transports = {};

// On initialize
const sessionId = uuidv4();
const transport = new StreamableHTTPServerTransport();
transports[sessionId] = transport;
await mcpServer.connect(transport);
res.setHeader('Mcp-Session-Id', sessionId);
await transport.handleRequest(req, res, body);

// On subsequent requests
const sessionId = req.headers['mcp-session-id'];
const transport = transports[sessionId];
await transport.handleRequest(req, res, body);
```

## Resources Found
- SimpleScraper guide: https://simplescraper.io/blog/how-to-mcp
- NapthaAI/http-oauth-mcp-server (uses proper transport pattern)
- AWS MCP implementation guide

## Next Steps
1. Implement transport-based request handling
2. Test initialize response includes tools
3. Verify tools appear in Claude.ai
4. Test tool invocation works

- [technical] Root cause: Not using MCP SDK transport classes
- [implementation] Need StreamableHTTPServerTransport for proper protocol handling
- [discovery] Tools exposed in initialize response, not tools/list