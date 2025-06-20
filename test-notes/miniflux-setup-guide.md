---
title: Miniflux Setup Guide for Portainer
date: 2025-05-26 19:18:32
tags:
  - miniflux
  - rss
  - selfhosted
modified: 2025-05-26 21:04:05
permalink: basic-memory/miniflux-setup-guide
---

# Miniflux Setup Guide for Portainer

## 1. Deploy in Portainer

### Option A: Stack Deploy (Recommended)

1. In Portainer, go to **Stacks** → **Add Stack**
2. Name it: `miniflux`
3. Paste the docker-compose content
4. Update these values before deploying:
   - `ADMIN_PASSWORD`: Choose a strong password
   - `POSTGRES_PASSWORD`: Choose a different strong password
   - `BASE_URL`: Your domain (or use http://nas-ip:8080)
   - Port `8080`: Change if needed

### Option B: Individual Containers

1. Create network `miniflux-network` first
2. Deploy PostgreSQL container
3. Deploy Miniflux container with env vars

## 2. Initial Setup

1. Access Miniflux at `http://your-nas:8080`
2. Login with:
   - Username: `admin`
   - Password: (what you set in ADMIN_PASSWORD)

3. Change admin password immediately:
   - Settings → Password

## 3. Import Your Feeds

1. Export from Feedly:
   - Feedly → Organize → Export OPML

2. Import to Miniflux:
   - Feeds → Import → Choose OPML file

3. Or import your generated OPML:

   ```bash
   ~/bin/generate_rss_urls.rb opml > mastodon_feeds.opml
   ```

## 4. Set Up Filtering Rules

Go to **Settings** → **Rules** and add:

### Global Block Rules (like your stopwords):

```
# Block political content
(?i)(breaking|election|politics|trump|biden|congress|ukraine|russia)

# Block negativity
(?i)(outrage|angry|rant|hate|toxic|problematic)

# Block tech drama
(?i)(elon|musk|twitter|drama|hot.take)
```

### Per-Feed Rules:

1. Click on a feed
2. Go to "Edit" 
3. Add rules specific to that feed

### Rule Types:

- **Block rules**: Hide entries matching pattern
- **Keep rules**: Only show entries matching pattern
- **Block/Keep based on**: Title, Content, Author, URL

## 5. Keyboard Shortcuts (Vim-like!)

- `g u`: Go to unread
- `g b`: Go to bookmarks  
- `j/k`: Next/previous item
- `v`: Open original link
- `m`: Toggle read status
- `f`: Star/bookmark

## 6. API for Automation

Get your API key:
- Settings → API Keys → Create Key

Then you can automate adding feeds:

```bash
# Add a feed via API
curl -X POST https://your-miniflux/v1/feeds \
  -H "X-Auth-Token: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"feed_url": "https://mastodon.social/@someone.rss", "category_id": 1}'
```

## 7. Mobile Apps

- **iOS**: Reeder, NetNewsWire, Unread
- **Android**: FeedMe, Reeder
- All support Miniflux's Fever API

## 8. Tips

- Set up categories matching your Mastodon lists
- Use "Refresh all feeds" sparingly (respects rate limits)
- Enable "Fetch original content" for truncated feeds
- Set polling frequency based on your needs (default 60 min is good)

## Migrating from Feedly

The best part: Miniflux respects your reading habits:
- Starred items → Bookmarks
- Categories transfer over
- Read/unread status maintained
- No ads, no limits, no algorithm

Your Miniflux instance becomes your personal, filtered view of the creative web!
