---
date: 2025-06-12 22:51:28
tags:
title: Things MCP - List Headings Implementation Summary
type: note
permalink: basic-memory/things-mcp-list-headings-implementation-summary
modified: 2025-06-13 07:32:16
---

# Things MCP - List Headings Implementation Summary

## Problem Solved

You hit a wall with adding items under headings because users didn't know what headings existed in projects. The `heading` parameter in `create_task` works, but only if the heading exists and is named exactly right.

## Solution Implemented

Added comprehensive heading support with enumeration and validation:

### 1. New Tool: `list_headings_by_project`

- **Purpose**: Enumerate all headings in a specific project
- **Usage**: `list_headings_by_project(project_name: "Project Name")`
- **Detection**: Uses heuristic to identify headings (items with no tags, dates, or checklists)
- **Output**: Shows heading names, IDs, and usage example

### 2. Enhanced Validation: `validate_heading_exists`

- **Purpose**: Validate heading exists before creating tasks
- **Integration**: Automatically called by `create_task` when heading parameter is used
- **Error Messages**: 
  - Lists available headings when invalid heading specified
  - Requires project when heading specified without project
  - Helpful guidance to use `list_headings_by_project` tool

### 3. Implementation Details

- **Heuristic Detection**: Items are identified as headings if they have:
  - No tags
  - No due date
  - No activation date  
  - No checklist items
  - Status is 'open'
- **Heading IDs**: Each heading has a unique ID that can be used with `heading-id` parameter
- **Validation**: Prevents silent failures by checking heading existence before URL construction

## Test Results

✅ **Heading enumeration works**: Found 3 headings in "Things MCP" project
✅ **Validation works**: Properly catches invalid headings and provides helpful errors
✅ **Existing functionality preserved**: `list_tasks_by_project` still works correctly
✅ **Error handling**: Clear messages guide users to use enumeration tool first

## Usage Workflow

1. **Discover headings**: `list_headings_by_project("Project Name")`
2. **Create tasks under headings**: `create_task(name: "Task", project: "Project", heading: "Heading Name")`
3. **Validation happens automatically**: No silent failures

## Files Modified

- `things-mcp.rb`: Added tool definition, handler, implementation, and validation
- Test files created: `test-manual-headings.rb`, `test-validation.rb`

## API Insight

The Things URL scheme supports both `heading` (by name) and `heading-id` (by ID) parameters, but both are silently ignored if the heading doesn't exist. The validation prevents this silent failure mode.

This resolves the "wall" you encountered with heading functionality!
