---
title: Pixelfed Enhancement Ideas for imgup-cli
type: note
permalink: basic-memory/pixelfed-enhancement-ideas-for-imgup-cli
tags:
- '#imgup'
- '#pixelfed'
- '#fediverse'
- '#feature-ideas'
- '#photography'
---

# Pixelfed Enhancement Ideas for imgup-cli

## Context
During development of fediverse support for imgup-cli, we discussed potential Pixelfed-specific features. While Pixelfed works with the existing Mastodon API implementation, it has unique photo-focused features that could be leveraged.

## Current Status
- Pixelfed already works with `imgup setup mastodon` (just use Pixelfed instance URL)
- Uses same Mastodon API for basic posting
- No Pixelfed-specific features implemented yet

## Potential Pixelfed Enhancements

### 1. Albums/Collections Support
- Add `--album "Album Name"` flag to group photos
- Would use Pixelfed's collection API to create/add to albums
- Example:
  ```bash
  imgup -b smugmug --pixelfed \
    --album "Oregon Coast 2025" \
    --image sunset.jpg --desc "Golden hour" \
    --image lighthouse.jpg --desc "Heceta Head"
  ```

### 2. Stories (Ephemeral Posts)
- Add `--story` flag for 24-hour posts
- Uses different endpoint: `/api/v1/stories`
- Example:
  ```bash
  imgup -b flickr --pixelfed --story \
    --image concert.jpg --desc "Amazing show tonight!"
  ```

### 3. Location/Place Tags
- Add `--location "City, State"` option
- Extract GPS coordinates from EXIF data
- Pixelfed displays location prominently in UI
- Uses `place_id` in post creation

### 4. Creative Commons Licensing
- Add `--license` flag with options:
  - `cc-by` - Attribution
  - `cc-by-sa` - Attribution ShareAlike
  - `cc-by-nc` - Attribution NonCommercial
  - `cc-by-nc-sa` - Attribution NonCommercial ShareAlike
  - `cc-by-nd` - Attribution NoDerivatives
  - `cc-by-nc-nd` - Attribution NonCommercial NoDerivatives
  - `cc0` - Public Domain
- Important for photographers who want proper attribution

### 5. Enhanced Content Warnings
- More granular than Mastodon's binary sensitive flag
- Add `--content-warning "Description"` for specific warnings
- Useful for art photography, journalism, etc.

## Implementation Notes
- Would need to detect if instance is Pixelfed (check API endpoints)
- Could add `imgup setup pixelfed` for clearer UX
- Some features (albums, stories) require different API endpoints
- Would need Pixelfed instance for testing

## Why This Matters
Pixelfed is specifically designed for photographers, making it potentially the most natural fediverse platform for a photo upload tool. These features would make imgup a more complete Pixelfed client rather than just treating it as "another Mastodon instance."