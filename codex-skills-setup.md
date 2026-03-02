# Codex Skills Setup Guide

This repo stores custom Codex skills so they can be versioned in git and reused on any machine.

## Source of truth

- Repo path:
  - `.codex/skills/feature`
  - `.codex/skills/story`

## Important runtime detail

Codex loads skills from `~/.codex/skills`, not directly from the project folder.

Because of that, each machine needs a one-time setup step.

## Recommended setup (symlink)

Run from the repo root:

```bash
mkdir -p ~/.codex/skills
ln -sfn "$(pwd)/.codex/skills/feature" ~/.codex/skills/feature
ln -sfn "$(pwd)/.codex/skills/story" ~/.codex/skills/story
```

Why this is recommended:
- Repo stays the single source of truth
- Changes in repo are instantly reflected in Codex runtime path
- No repeated copy/sync needed after each edit

## Alternative setup (copy)

If symlinks are not desired:

```bash
mkdir -p ~/.codex/skills
cp -r .codex/skills/feature ~/.codex/skills/
cp -r .codex/skills/story ~/.codex/skills/
```

Note: with copy mode, re-run copy commands whenever skills change.

## Verify setup

```bash
ls -la ~/.codex/skills/feature
ls -la ~/.codex/skills/story
```

## After setup

- Restart Codex so it reloads skills.
- Use skills explicitly via:
  - `$feature ...`
  - `$story ...`

## Team workflow

1. Edit skills in repo under `.codex/skills/...`
2. Commit changes to git
3. On another machine, pull latest repo and run one-time setup commands above
