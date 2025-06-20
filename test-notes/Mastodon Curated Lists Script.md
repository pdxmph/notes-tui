---
date: 2025-05-26 15:52:03
title: Mastodon Curated Lists Script
type: note
permalink: basic-memory/mastodon-curated-lists-script
tags: [mastodon, fediverse, automation, ruby, social-media, photography, filmphotography]
modified: 2025-06-12 22:32:49
---

# Mastodon Curated Lists Script

## Overview

`curated_lists.rb` is a Ruby script that helps create curated Mastodon lists for "Exposure Therapy" - discovering creative content without news/politics/hot takes cluttering your timeline.

## What It Does

1. **Discovers creative accounts** from Fediverse instances known for photography/writing
2. **Filters out noise** using configurable stopwords (politics, news, etc.)
3. **Follows accounts with boosts hidden** (`reblogs: false`) to keep main timeline clean
4. **Adds accounts to themed lists** (Photography Inspiration, Writing & Poetry, etc.)
5. **Tracks progress** to avoid duplicate processing

## Key Features

- Checks if you already follow accounts (won't create duplicate follows)
- Rate limits API calls (2 second delay between actions)
- Asks for confirmation before running
- Processes a configurable number of accounts per run

## Files Used

### Input Files

- `~/.config/mastodon_discovery/config.yml` - Main configuration:
  - Mastodon instance and access token
  - List definitions (Photography Inspiration, Writing & Poetry, Creative Process)
  - Discovery sources (instances and hashtags to search)
  - Stopwords for filtering (politics, news, negativity terms)
  - Settings (minimum followers, rate limits)

### Output Files  

- `~/.config/mastodon_discovery/processed.json` - Tracking file:
  - List of already-processed accounts
  - Statistics by category

## Usage

```bash
# Process 5 accounts (good for testing)
~/bin/curated_lists.rb 5

# Process 10 accounts (default)
~/bin/curated_lists.rb

# Process 20 accounts
~/bin/curated_lists.rb 20
```

## How It Works

1. Searches hashtags on instances like photog.social, mastodon.art
2. Analyzes posts to find accounts that:
   - Post images (for photography category)
   - Don't mention politics/news keywords
   - Have minimum follower threshold (default: 25)
3. For each qualifying account:
   - Checks if you already follow them
   - If not following: Follows with boosts hidden
   - Adds to appropriate list
   - Records in tracking file

## Configuration

Edit `~/.config/mastodon_discovery/config.yml` to:
- Add/remove stopwords
- Change minimum follower threshold  
- Add/remove hashtags to search
- Add/remove instances to check

## Result

- **Main timeline**: Stays clean (no boosts from curated accounts)
- **Lists**: Become focused inspiration feeds
- **Morning routine**: Check lists for creative inspiration without 

## Technical Notes

- Uses Mastodon API v1 endpoints
- Requires token with scopes: read, write:lists, write:follows
- Works with social.lol's requirement that you must follow to add to lists
- Handles API errors gracefully
- Saves state between runs
