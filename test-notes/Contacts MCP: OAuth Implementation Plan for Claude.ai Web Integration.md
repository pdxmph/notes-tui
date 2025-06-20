---
title: 'Contacts MCP: OAuth Implementation Plan for Claude.ai Web Integration'
type: note
permalink: basic-memory/contacts-mcp-oauth-implementation-plan-for-claude-ai-web-integration
---

# Contacts MCP: OAuth Implementation Plan for Claude.ai Web Integration

## Current Status (Working)

**✅ Phase 1 Complete - Claude Desktop Integration**
- **Repository**: `pdxmph/contact-mcp-ts`
- **Deployment**: `https://contact-mcp-ts.mike-7b5.workers.dev`
- **Working Features**:
  - MCP-over-HTTP (JSON-RPC) at `/mcp` endpoint
  - Bearer token authentication (`contact-api-secret-123`)
  - D1 database with 229 contacts migrated
  - 4 core tools working: `search_contacts`, `add_contact`, `get_contact`, `mark_contacted`
  - Claude Desktop successfully configured and operational

**✅ Claude Desktop Configuration (Working)**
```json
{
  "mcpServers": {
    "contacts-cloudflare": {
      "url": "https://contact-mcp-ts.mike-7b5.workers.dev/mcp",
      "auth": {
        "type": "bearer",
        "token": "contact-api-secret-123"
      }
    }
  }
}
```

---

## Phase 2 Goal: Claude.ai Web Integration

**Objective**: Enable contact access via Claude.ai web interface for cross-device usage

**Requirements Discovered**:
- Claude.ai web only supports MCP-over-SSE (Server-Sent Events)
- Authentication options: Authless OR OAuth with Dynamic Client Registration
- No bearer token support for Claude.ai web integrations
- Personal use only (not multi-user system)

**User Requirements**:
- ✅ **GitHub OAuth** for proper authentication (no security through obscurity)
- ✅ **Keep bearer token** for Claude Desktop (fallback strategy)
- ✅ **Personal authorization** - only mph's GitHub account gets access

---

## Technical Architecture

### Dual Authentication System

**Endpoint Strategy**:
- `/mcp` - Bearer token + JSON-RPC (existing, Claude Desktop)
- `/mcp/oauth` - GitHub OAuth + SSE (new, Claude.ai web)

**Transport Summary**:
- **Claude Desktop**: MCP-over-HTTP (JSON-RPC) with bearer token
- **Claude.ai Web**: MCP-over-SSE with GitHub OAuth

### OAuth Flow Design

**1. Authentication Endpoints Required**:
- `GET /authorize` - Start OAuth flow, redirect to GitHub
- `POST /token` - Exchange authorization code for access token
- `GET /callback` - Handle GitHub OAuth callback
- `GET /.well-known/oauth-authorization-server` - OAuth metadata discovery

**2. OAuth Flow Steps**:
1. Claude.ai redirects user to our `/authorize` endpoint
2. Our server redirects to GitHub OAuth with appropriate scopes
3. User authenticates with GitHub
4. GitHub redirects to our `/callback` with authorization code
5. Our server exchanges code for GitHub access token
6. Our server validates: "Is this mph's GitHub account?" (hardcoded check)
7. If valid: generate MCP access token and complete OAuth flow
8. Claude.ai establishes authenticated SSE connection to `/mcp/oauth`

**3. GitHub OAuth Setup Required**:
- Create GitHub OAuth app in pdxmph account
- Configure callback URL: `https://contact-mcp-ts.mike-7b5.workers.dev/callback`
- Store client ID and secret in Cloudflare environment variables

### Authorization Logic

**Personal Authorization Strategy**:
- Hardcode mph's GitHub username/ID in environment variable
- OAuth flow validates GitHub identity against this value
- Only matching GitHub account gets contact access
- No complex user management needed

**Fallback Strategy**:
- Bearer token endpoint (`/mcp`) remains unchanged
- If OAuth implementation fails, Claude Desktop continues working
- Can fall back to authless SSE as emergency option

---

## Implementation Requirements

### Code Changes Needed

**1. Worker Updates (`src/worker.ts`)**:
- Add SSE transport support using `@modelcontextprotocol/sdk/server/sse.js`
- Implement OAuth endpoints (`/authorize`, `/token`, `/callback`)
- Add OAuth metadata discovery endpoint
- Create authenticated SSE handler at `/mcp/oauth`
- Maintain existing JSON-RPC handler at `/mcp`

**2. Environment Variables**:
- `GITHUB_CLIENT_ID` - GitHub OAuth app client ID
- `GITHUB_CLIENT_SECRET` - GitHub OAuth app secret  
- `AUTHORIZED_GITHUB_USER` - mph's GitHub username for validation
- `BEARER_TOKEN` - existing token (keep for Claude Desktop)

**3. Dependencies**:
- May need additional OAuth libraries for token handling
- Ensure SSE transport properly configured

### Infrastructure Setup

**1. GitHub OAuth App**:
- Register OAuth app in pdxmph GitHub account
- Application name: "Contacts MCP Server"
- Homepage URL: `https://contact-mcp-ts.mike-7b5.workers.dev`
- Callback URL: `https://contact-mcp-ts.mike-7b5.workers.dev/callback`
- Required scopes: `user:email` (to identify user)

**2. Cloudflare Configuration**:
- Add environment variables via `wrangler secret put`
- Update `wrangler.toml` if needed for new endpoints
- Test SSE compatibility with Cloudflare Workers

### Testing Strategy

**1. OAuth Flow Testing**:
- Test complete GitHub OAuth flow manually
- Verify only authorized GitHub user gets access
- Test unauthorized users get proper rejection
- Validate token generation and validation

**2. SSE Transport Testing**:
- Test SSE connection establishment
- Verify MCP tools work over SSE
- Test with MCP Inspector tool
- Validate Claude.ai web integration

**3. Fallback Testing**:
- Ensure Claude Desktop continues working during OAuth implementation
- Test both endpoints simultaneously
- Verify no regression in existing functionality

---

## Next Steps for Implementation

### Issue Creation Strategy

**Issue 1: GitHub OAuth Infrastructure Setup**
- Create GitHub OAuth app
- Configure environment variables
- Add OAuth endpoints (basic structure)
- Test OAuth flow without MCP integration

**Issue 2: SSE Transport Implementation**  
- Add SSE transport to worker
- Create `/mcp/oauth` endpoint
- Integrate with OAuth authentication
- Test SSE connection and tool execution

**Issue 3: OAuth Metadata & Discovery**
- Implement `/.well-known/oauth-authorization-server`
- Add Dynamic Client Registration support if needed
- Ensure Claude.ai web can discover OAuth capabilities

**Issue 4: Integration Testing & Deployment**
- Test complete flow with Claude.ai web
- Validate both Claude Desktop and web work simultaneously
- Performance testing and optimization
- Documentation updates

### Success Criteria

**Minimum Success**:
- GitHub OAuth flow working
- SSE transport functional  
- Claude.ai web can authenticate and use contacts
- Claude Desktop continues working unchanged

**Full Success**:
- Seamless Claude.ai web integration
- Proper security (only mph's GitHub account)
- Reliable fallback to Claude Desktop
- Performance acceptable for both interfaces

---

## Technical References

**Documentation Links**:
- [MCP Auth Specification](https://modelcontextprotocol.io/specification/2025-03-26/basic/authorization)
- [Claude.ai Custom Integrations](https://support.anthropic.com/en/articles/11503834-building-custom-integrations-via-remote-mcp-servers)
- [OAuth 2.1 Draft Specification](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1-12)
- [Cloudflare Workers D1 Documentation](https://developers.cloudflare.com/d1/)

**Repository Context**:
- **Current working branch**: main
- **Key files**: `src/worker.ts`, `src/database.ts`, `wrangler.toml`
- **Database**: D1 with contacts, contact_interactions, logs tables
- **Build command**: `npm run build:worker`
- **Deploy command**: `npx wrangler deploy`

---

## Implementation Notes

**Architecture Decision Rationale**:
- Dual endpoints avoid breaking existing Claude Desktop setup
- GitHub OAuth provides proper authentication without multi-user complexity
- Personal authorization (hardcoded user check) keeps implementation simple
- SSE transport meets Claude.ai web requirements
- Fallback strategy ensures reliability

**Security Considerations**:
- No security through obscurity (per user requirement)
- Proper OAuth flow with GitHub identity validation
- Access tokens with appropriate expiration
- HTTPS required for all OAuth endpoints
- Secure storage of GitHub OAuth credentials

**Future Enhancements Possible**:
- Additional OAuth providers (Google, etc.)
- Multi-user support (if ever needed)
- Additional MCP tools (expand from 4 to 24 available)
- Advanced contact management features

---

*Status: Ready for implementation planning and issue creation*  
*Created: June 15, 2025*