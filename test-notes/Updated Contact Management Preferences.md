---
title: Updated Contact Management Preferences
type: note
permalink: basic-memory/updated-contact-management-preferences
---

# Contacts 

I maintain a contacts database with relationship tracking integrated with Basic Memory for rich context. When I mention people:

- Check if they exist in my contacts system using search tools
- Reference their relationship type (work, close, family, network, etc.)
- If they have a basic_memory_url, pull their Basic Memory note for full context using build_context
- Consider if contact records need updates (last contacted, interaction notes)
- When updating contact information, sync changes to their Basic Memory note
- Suggest when to follow up on overdue relationships
- Use contact context to inform work planning and task assignments

Key contacts have a "label" field to make it easier to find frequently referenced people and disambiguate between similarly named people. Labels follow the @firstnamelastinitial convention.

## Information Architecture - Three Storage Layers

### 1. Database `notes` Field - Brief Status Only
- **Purpose**: Quick identification and role context
- **Format**: One-liner status (e.g., "CISO - Mike's manager", "Frontend dev at Stripe")
- **Usage**: Fast database queries, contact list display
- **Update**: Only when role/status fundamentally changes

### 2. Database Logs - Automated Interaction Tracking
- **Purpose**: Chronological contact history with basic facts
- **Format**: Timestamped entries with interaction type and brief context
- **Usage**: "When did I last contact X?", automated email sync, contact frequency
- **Update**: Automated via tools, manual for significant interactions
- **Content**: Dates, interaction types, brief factual context only

### 3. Basic Memory Notes - Rich Narrative Context
- **Purpose**: All qualitative observations, relationship insights, detailed context
- **Format**: Structured markdown with background, observations, action items
- **Usage**: Meeting prep, relationship building, AI collaboration, deep context
- **Update**: All qualitative observations go here via `update_contact_note`
- **Content**: Personality insights, working styles, relationship dynamics, strategic context

## Contact Information Workflow

### When Interacting with Contacts:
1. **Basic interaction tracking** → Update database log (date, type, brief facts)
2. **Qualitative observations** → Add to Basic Memory note via `update_contact_note`
3. **Role/status changes** → Update database `notes` field
4. **AI conversations about people** → Pull full context from Basic Memory notes

### When Someone is Mentioned:
1. **Search contacts database** for basic info and relationship type
2. **If basic_memory_url exists** → Use `build_context` to load rich narrative
3. **Reference full context** when providing advice or insights
4. **Suggest updates** if new information emerges during conversation

### For New Qualitative Observations:
- **Always use Basic Memory notes** (not database logs or notes field)
- **Use `update_contact_note`** to sync observations to Basic Memory
- **Keep database logs factual only** (when, what type, brief context)
- **Preserve the semantic distinction** between facts and insights

This architecture keeps the database fast for queries while centralizing all rich relationship context in Basic Memory for AI collaboration, similar to the original org-mode workflow but in markdown format.

# Basic Memory Contact Note Template

## Note Format

Contact notes in Basic Memory use YAML frontmatter for structured metadata and markdown content for rich context:

```markdown
---
contact_id: [database_id]
relationship: [type]
company: [company]
last_contact: [YYYY-MM-DD]
state: [state]
label: "[label]"
email: [email]
phone: [phone]
---

# Contact Name (@label)

## Background
[Rich context about relationship, how you know them, career history, personal details]

## Recent Interactions
[Chronological interaction history with dates and context]

## Observations
- [category] Observation about working style, personality, expertise (timestamp)
- [category] Professional insights and behavioral patterns (timestamp)

## Relationships
- [category] How they relate to other people/teams/organizations (timestamp)
- [category] Organizational dynamics and network connections (timestamp)

## Action Items
- [ ] Next steps and follow-ups
- [ ] Introductions to make
- [ ] Projects to discuss

## Links
**Contact ID:** [database_id]  
**LinkedIn:** [url]  
**Related contacts:** [other contacts]  
**Projects:** [JIRA tickets, GitHub issues, etc.]
```

## YAML Fields

- `contact_id`: Database ID for linking to contacts MCP
- `relationship`: work, close, family, network, social, providers, recruiters
- `company`: Current employer/organization
- `last_contact`: Date of most recent interaction (YYYY-MM-DD format)
- `state`: Current contact state (ok, ping, invite, write, followup, etc.)
- `label`: Quick reference handle (quoted, e.g., "@jeffv")
- `email`: Primary email address
- `phone`: Phone number

## Observation Categories

Common observation types:
- `[leadership]` - Management and decision-making style
- `[technical]` - Technical expertise and approach
- `[communication]` - Communication preferences and style
- `[collaboration]` - How they work with others
- `[expertise]` - Domain knowledge and specializations
- `[personality]` - Character traits and behavioral patterns
- `[goals]` - Professional or personal objectives
- `[challenges]` - Current difficulties or obstacles

## Relationship Categories

Common relationship types:
- `[work]` - Professional reporting and collaboration relationships
- `[network]` - Professional network connections
- `[personal]` - Personal friendships and family connections
- `[project]` - Project-specific working relationships
- `[challenge]` - Problematic or difficult relationships
- `[opportunity]` - Potential collaboration or connection opportunities
- `[support]` - Mentoring or support relationships

## Contact Updates Protocol

When updating contact information or relationships:

1. Update the contacts database first
2. If contact has a basic_memory_url, update their Basic Memory note
3. Format observations in the standard Basic Memory style:
    - `- [observation type] Observation text (context)`
    - Types include: professional, personal, connection, interest, goal, challenge
4. Add tasks related to the person as markdown todos in their Tasks section
5. Ensure bidirectional sync between database and note

## Relationship Context

When referencing a contact in conversation:

- Always check for and load their Basic Memory note if available
- Use the full context from their note to inform responses
- Update the note with new information learned during our conversation
- Track relationship changes and patterns in the Observations section

## Contact Notes - Factual Documentation Only

When creating or updating contact notes:

- Document only the specific facts, interactions, and context provided
- Do NOT add qualitative assessments, character judgments, or interpretations
- Do NOT infer personality traits or behavioral patterns beyond what was explicitly stated
- Stick to observable actions and direct quotes/paraphrases from interactions
- Let the user draw their own conclusions from the documented facts

Example of what NOT to do:
- "Shows good judgment" 
- "Values proper consultation"
- "Comfortable speaking up"

Example of what TO do:
- "Reached out with concerns about meeting"
- "Said he felt 'steamrolled' into the meeting"
- "Meeting was canceled after discussion"