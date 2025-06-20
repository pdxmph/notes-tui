---
date: 2025-06-08 19:53:57
title: MCP Protocol Version Confusion
type: note
permalink: basic-memory/mcp-protocol-version-confusion
tags: ['mcp', 'protocol', 'debugging', 'claude-training']
modified: 2025-06-12 22:25:38
---

# MCP Protocol Version Confusion

## The Issue

When implementing MCP (Model Context Protocol) servers, there's a recurring confusion about the protocol version. Claude's training data appears to suggest the protocol version should be `"0.1.0"`, but the actual Claude Desktop client expects `"2024-11-05"`.

## Symptoms

When using the wrong protocol version, you'll see errors like:
- `"Unrecognized key(s) in object: 'error'"` 
- Multiple `ZodError` validation failures
- The MCP server disconnecting immediately after initialization

## The Solution

Always use protocol version `"2024-11-05"` in your MCP server's initialize response:

```json
{
  "jsonrpc": "2.0",
  "id": 0,
  "result": {
    "protocolVersion": "2024-11-05",  // NOT "0.1.0"!
    "capabilities": {
      "tools": {}
    },
    "serverInfo": {
      "name": "your-mcp-server",
      "version": "1.0.0"
    }
  }
}
```

## Additional Protocol Notes

1. **Notifications vs Requests**: Some messages like `notifications/initialized` are notifications (no ID field). Don't send responses to notifications.

2. **JSON-RPC Compliance**: Always include `"jsonrpc": "2.0"` in all messages.

3. **ID Handling**: The ID field can be a string, number, or null (for notifications).

## Why This Happens

This appears to be a training data issue where Claude's knowledge about MCP protocol versions doesn't match the current implementation in Claude Desktop. The protocol likely evolved after Claude's training cutoff, changing from semantic versioning (0.1.0) to date-based versioning (2024-11-05).

## Debugging Tips

When implementing an MCP server:
1. Log all incoming messages to see what the client expects
2. Check the protocol version in the client's initialize request
3. Mirror back the same protocol version the client sends
4. Use stderr for logging (stdout is reserved for JSON-RPC messages)

This issue has come up multiple times when implementing MCP servers, so always double-check the protocol version when debugging connection issues.
