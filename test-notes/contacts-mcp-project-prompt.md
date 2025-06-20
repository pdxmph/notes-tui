---
title: Contacts MCP Project Prompt
date: 2025-06-08 09:03:59
tags:
  - mcp
  - project-state
  - claude
  - prompts
  - cicd
modified: 2025-06-16 16:35:27
permalink: basic-memory/contacts-mcp-project-prompt
type: note
aliases: [Contacts MCP Project Prompt]
---

# Contacts MCP Project Prompt

## Project Overview

Continue development of the Contacts MCP (Model Context Protocol) server in Go that allows Claude to manage personal contacts via native MCP tools. This project builds a comprehensive contact relationship management system with intelligent categorization and maintenance tracking.

### 🚨 CRITICAL DEVELOPMENT RULES

1. **⚠️ ALL CONTACT OPERATIONS MUST RESPECT RELATIONSHIP CATEGORIES**
   - **NO EXCEPTIONS** when moving contacts between groups
   - Relationship types are: `close`, `family`, `network`, `social`, `providers`, `recruiters`, `work`
   - Different overdue thresholds: 30 days (close/family/social/work) vs 90 days (network/providers/recruiters)
   - **Social group serves as "holding tank"** for people you know but don't actively maintain
   - This categorization is **NON-NEGOTIABLE** for maintaining contact organization

2. **🎯 DATABASE IS THE SINGLE SOURCE OF TRUTH FOR CONTACTS**
   - SQLite database at `~/.contacts/contacts.db` contains authoritative contact data
   - All contact modifications go through MCP tools, not direct database access
   - When debugging, check database directly but modify through tools
   - **Database integrity is NON-NEGOTIABLE for contact management**

3. **🔄 MANDATORY RESTART PROTOCOL** 
   - **WHEN YOU IMPLEMENT NEW TOOLS OR MODIFY EXISTING ONES:**
     1. Complete the implementation
     2. **IMMEDIATELY STOP** all other activity
     3. Tell the user: "Implementation complete. **Claude Desktop must be restarted** to register the new tool. Please restart Claude Desktop and return to test the [tool name] functionality."
     4. **DO NOT** try to test the tool until after restart
     5. **DO NOT** continue with other tasks until restart is confirmed

   - **WHEN A TOOL IS MISSING FROM AVAILABLE TOOLS:**
     1. **IMMEDIATELY STOP** and ask: "I don't see the [tool name] tool available. Can you confirm if it appears in your MCP tool list?"
     2. If user confirms it's missing: "The tool needs a restart to register. Please restart Claude Desktop and return to test the [tool name] functionality."
     3. **DO NOT** assume the tool will work without confirmation

## ✅ CURRENT PROJECT STATUS (Updated 2025-06-10)

### 🎉 **MAJOR MILESTONES ACHIEVED**

1. **✅ Go Rewrite Complete** (Issue #2)
   - Eliminated Ruby 500ms-1s startup overhead 
   - Achieved <50ms startup time
   - All 14 MCP tools implemented with full feature parity
   - Connection pooling, unified queries, composite indexes

2. **✅ HTTP/SSE Remote Access** (Issue #8) 
   - HTTP transport layer implemented
   - Server-Sent Events (SSE) endpoint at `/sse`
   - Successfully deployed at `https://contacts.puddingtime.net/sse`
   - Proxied behind Synology with Cloudflare Tunnel

3. **✅ Bearer Token Authentication** (Issue #10)
   - Simple bearer token auth middleware implemented
   - Secure access for Web Claude integration
   - Environment variable configuration: `MCP_AUTH_TOKEN`
   - Standards-compliant Authorization header support

4. **✅ Remote MCP Access Working** (Issue #7)
   - Successfully configured with `mcp-remote` proxy
   - Claude Desktop working config:

     ```json
     "contacts-mcp-remote": {
       "command": "npx",
       "args": [
         "mcp-remote",
         "https://contacts.puddingtime.net/sse",
         "--auth-bearer",
         "wjXFrReUkZO0Z5fo0j4+hb5k+wd0EBQ/MluaebfsZ/s="
       ]
     }
     ```

5. **✅ GitHub Actions CI/CD Pipeline Complete** (Issue #13)
   - Automated build and deployment to GitHub Container Registry
   - Multi-stage Dockerfile with CGO support for SQLite
   - Production docker-compose configuration for Portainer
   - Zero manual binary builds required
   - Container registry approach: `ghcr.io/pdxmph/contacts-mcp:latest`

6. **✅ Log Functionality Complete** (Issue #5)
   - Personal CRM capabilities with `add_log`, `list_logs`, `search_logs`
   - @mention parsing for linking logs to contacts
   - Full interaction history tracking
   - Search across all log content

7. **✅ Database Configuration** (Issue #4, #6)
   - Configurable database path via `CONTACTS_DB_PATH` environment variable
   - Automatic fallback paths for robust deployment
   - Schema migrations and data integrity

### 🔧 **DEPLOYED ARCHITECTURE**

**Local Development:**
- Go binary at `/Users/mph/code/contacts-mcp/contacts-mcp`
- SQLite database at `~/.contacts/contacts.db`
- Direct MCP connection via Claude Desktop

**Remote Production:**
- HTTP server with SSE transport
- Bearer token authentication
- Deployed on Synology via Portainer
- Container images automatically built via GitHub Actions
- Accessible at `https://contacts.puddingtime.net/sse`
- Works with Claude Desktop via `mcp-remote` proxy

**CI/CD Pipeline:**
- GitHub Actions builds Docker images on every commit to main
- Multi-stage Dockerfile with CGO enabled for SQLite support
- Images pushed to GitHub Container Registry (public package)
- Portainer pulls from `ghcr.io/pdxmph/contacts-mcp:latest`
- Zero-friction deployment: commit code → automatic build → update container

### ✅ Working Features

1. **Core Contact Management** (18 total tools)
   - `search_contacts` - Search contacts by name, email, company, notes, or label
   - `get_contact` - Get full contact details including interaction history
   - `add_contact` - Add new contacts with relationship categorization
   - `update_contact_info` - Update name, email, phone, company, notes, label
   - `change_contact_group` - Move contacts between relationship groups
   - `delete_contact` - Permanently remove contacts (with confirmation)

2. **Relationship Management**
   - `list_overdue_contacts` - List contacts past due thresholds by relationship type
   - `list_active_contacts` - List contacts with active TODO states
   - `report_by_group` - Generate comprehensive reports by relationship group
   - `mark_contacted` - Update contact date and clear overdue status
   - `bulk_mark_contacted` - Mark multiple contacts as contacted
   - `set_contact_state` - Set TODO states (ping, invite, write, followup, etc.)

3. **Contact Intelligence**
   - `get_contact_stats` - Overview statistics about contact management
   - `suggest_tasks` - Generate task suggestions for overdue contacts
   - **Intelligent NULL date handling** - Treats never-contacted as infinitely overdue
   - **Relationship-based thresholds** - 30 vs 90 day thresholds by group type

4. **Personal CRM / Logging**
   - `add_log` - Create log entries with @mention parsing
   - `list_logs` - List logs (filterable by contact, timeframe)
   - `search_logs` - Search log content
   - `get_contact_logs` - Get all logs for a specific contact

5. **Database Features**
   - **Label field support** - GTD-style @labels for easy reference and disambiguation
   - **Interaction tracking** - Full history of contact interactions
   - **Schema migrations** - Evolve schema without data loss
   - **Performance optimizations** - Connection pooling, composite indexes, unified queries

### 🔧 Key Implementation Details

- **Language:** Go (originally Ruby, rewritten for performance)
- **Database Location:** `~/.contacts/contacts.db` with configurable path
- **Relationship Types:** 7 categories with appropriate overdue thresholds
- **MCP Integration:** All operations through Claude-native tools
- **Performance:** Sub-50ms startup time, connection pooling, optimized queries
- **Remote Access:** HTTP/SSE transport with bearer token authentication
- **Data Integrity:** Foreign key constraints and proper transaction handling
- **Natural Language:** Due dates and contact states use friendly language
- **CI/CD:** GitHub Actions with multi-stage Docker builds and registry deployment

### 📂 Project Structure

```
/Users/mph/code/contacts-mcp/
├── contacts-mcp.go           # Main MCP server implementation
├── contacts/
│   ├── models.go            # Data structures and relationship configs
│   ├── database.go          # Optimized database operations
│   └── handlers.go          # All tool implementations (18 tools)
├── .github/workflows/
│   └── deploy.yml           # GitHub Actions CI/CD pipeline
├── Dockerfile.cicd          # Multi-stage build with CGO support
├── docker-compose.prod.yml  # Production Portainer configuration
├── CICD.md                  # CI/CD setup and deployment guide
├── Makefile                 # Build configuration
├── test-build.sh           # Build and startup time verification
└── ~/.contacts/
    └── contacts.db         # SQLite database (auto-created)
```

### 🚀 Running the MCP

**Local (Claude Desktop):**

```bash
# Build the Go binary
cd ~/code/contacts-mcp
make build

# In Claude Desktop MCP settings:
{
  "contacts-mcp": {
    "command": "/Users/mph/code/contacts-mcp/contacts-mcp"
  }
}
```

**Remote (via mcp-remote proxy):**

```json
{
  "contacts-mcp-remote": {
    "command": "npx",
    "args": [
      "mcp-remote", 
      "https://contacts.puddingtime.net/sse",
      "--auth-bearer",
      "wjXFrReUkZO0Z5fo0j4+hb5k+wd0EBQ/MluaebfsZ/s="
    ]
  }
}
```

**Production Deployment:**
- Code commits automatically trigger GitHub Actions build
- New images available at `ghcr.io/pdxmph/contacts-mcp:latest`
- Update Portainer stack to pull latest image
- Uses `docker-compose.prod.yml` configuration

### 💡 Design Principles

**Contact Management Philosophy:**
- **Social group as "holding tank"** - Elegant solution for people you know but don't actively maintain
- **Relationship-based thresholds** - Different expectations for different relationship types
- **NULL date handling critical** - Must treat never-contacted as overdue, not current
- **Batch operations essential** - Single-contact operations don't scale for hundreds of contacts
- **Personal CRM approach** - Log observations and interactions for relationship context

**Technical Decisions:**
- **Go over Ruby** - Eliminated 500ms-1s startup overhead per operation
- **HTTP/SSE transport** - Enables remote access from any Claude instance
- **Bearer token auth** - Simple, secure authentication for remote access
- **Connection pooling** - Persistent SQLite connections for better performance
- **Unified queries** - Single optimized query instead of 7 separate queries for overdue contacts
- **Composite indexes** - Strategic indexes for common query patterns
- **WAL mode** - Better concurrent access for SQLite
- **GitHub Actions CI/CD** - Professional deployment pipeline eliminates manual builds
- **Container registry approach** - Public packages for easy deployment

### 🎯 **NEXT DEVELOPMENT PRIORITIES**

**Immediate Testing Opportunities:**
1. **Web Claude Integration Testing** - Verify bearer token auth works with Web Claude custom integrations
2. **Performance monitoring** - Add logging/metrics for remote service usage
3. **End-to-end validation** - Test all 18 tools through remote connection

**Future Feature Enhancements:**
1. **Gmail API Integration** (Issue #11) - Auto-update contact dates from email interactions
2. **Auto-discover Email Contacts** (Issue #12) - Automatically add new contacts from email history
3. **Contact deduplication tools** - Find and merge duplicate contacts
4. **Calendar integration** - Track meetings as contact interactions
5. **Contact export** - Generate formatted contact lists for external use
6. **Advanced filtering** - More sophisticated contact queries and reports
7. **Contact notes enhancement** - Rich note-taking with timestamps and categories

**Longer-term Vision:**
1. **Voice interface integration** - Much better UX for rapid contact management
2. **AI-powered categorization** - Suggest relationship types based on contact patterns
3. **Social media integration** - Pull in LinkedIn/other social data
4. **Contact lifecycle management** - Automated relationship maintenance suggestions
5. **Mobile companion** - Sync with mobile contact management
6. **Team/organization support** - Multi-user contact management with permissions

### 🎊 **PROJECT SUCCESS METRICS**

- ✅ **Performance Goal**: <50ms startup (achieved vs 500ms-1s Ruby)
- ✅ **Remote Access Goal**: Working from any Claude instance
- ✅ **Feature Parity Goal**: All Ruby functionality replicated in Go
- ✅ **Security Goal**: Authenticated remote access
- ✅ **Deployment Goal**: Production-ready service on Synology
- ✅ **Personal CRM Goal**: Log functionality for relationship context
- ✅ **CI/CD Goal**: Zero-friction deployment pipeline
- ✅ **Infrastructure Goal**: Professional containerized deployment

**Ready for production use and advanced feature development!**
