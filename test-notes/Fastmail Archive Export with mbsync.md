---
date: 2025-05-25 12:58:13
title: Fastmail Archive Export with mbsync
type: note
permalink: basic-memory/fastmail-archive-export-with-mbsync
tags: [email archive fastmail mbsync notmuch backup]
modified: 2025-06-12 22:25:30
---

# Fastmail Archive Export with mbsync

## Context

Exporting 22+ years of email (2001-2023) from Fastmail for archival with notmuch. Running on Mac Studio using tmux for long-running sync.

## mbsync Configuration

### Installation

```bash
brew install isync
```

### Password Setup

```bash
# Create password file
echo "your-fastmail-app-password" > ~/.mbsync-fastmail-pass
chmod 600 ~/.mbsync-fastmail-pass
```

Note: Use Fastmail app password from Settings → Password & Security → App Passwords

### ~/.mbsyncrc Configuration

```
IMAPAccount fastmail
Host imap.fastmail.com
Port 993
User your-username@fastmail.com
PassCmd "cat ~/.mbsync-fastmail-pass"
SSLType IMAPS
AuthMechs LOGIN
Timeout 120
PipelineDepth 50

IMAPStore fastmail-remote
Account fastmail

MaildirStore fastmail-local
Path ~/Mail/fastmail/
Inbox ~/Mail/fastmail/INBOX
SubFolders Verbatim

Channel fastmail
Far :fastmail-remote:
Near :fastmail-local:
Patterns *
Create Near
Expunge Near
SyncState *
MaxMessages 0
```

## tmux Session Management

### Start sync

```bash
tmux new -s fastmail-export
mbsync -V fastmail
```

### Detach/Reattach

- **Detach**: `Ctrl-b d`
- **Reattach**: `tmux attach -t fastmail-export` or `tmux a -t fastmail-export`
- **List sessions**: `tmux ls`

### Monitor Progress

```bash
# Check disk usage
du -sh ~/Mail/fastmail/

# Count synced messages
find ~/Mail/fastmail -name "*" -type f | wc -l

# Keep Mac awake if needed
caffeinate -i tmux attach -t fastmail-export
```

## Next Steps

1. Complete sync (may take days for 22 years)
2. Set up notmuch indexing
3. Configure backup strategy:
   - Syncthing to NAS and other machines
   - Arq backup to NAS and Backblaze/Glacier
   - Exclude `.notmuch/` from syncs
