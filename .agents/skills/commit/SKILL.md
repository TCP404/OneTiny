---
name: commit
description: Generate commit messages following Conventional Commits format and commit staged changes. Use when the user explicitly asks to commit, create a commit, or run /commit.
argument-hint: "[message advice]"
disable-model-invocation: true
allowed-tools: Bash(rm -f ./.git/index.lock)
---

# Commit Skill

Generate well-formatted commit messages and commit staged changes using git.

## Usage

- `/commit` - Generates a commit message based on staged changes
- `/commit <advice>` - Uses provided advice to guide commit message generation

## Guidelines

- **Only commit staged files** - Never add files with `git add`. The user controls staging.
- **Analyze all staged changes** - Review both previously staged and newly added changes
- **Follow commit message format** - Use Conventional Commits format shown below
- **Match the project style** - Check recent commits (`git log`) to match existing style

## Format

```
<type>[!](<scope>): <message title>

<bullet points summarizing changes>

[optional BREAKING CHANGE section if applicable]
```

## Examples

### Basic commits

```
Feat(auth): add JWT login flow

- Implemented JWT token validation logic
- Added documentation for validation component
```

```
Fix(ui): handle null pointer in sidebar

- Added null check before accessing user object
- Updated error handling to show fallback UI
```

```
Refactor(api): split user controller logic

- Extracted user validation into separate module
- Simplified controller methods
```

### Breaking change

```
Chore!: drop support for Node 6

- Dropped support for older JavaScript syntax
- Updated dependencies to latest versions

BREAKING CHANGE: use JavaScript features not available in Node 6.
```

## Rules

- **Title**: Capitalized, no period, max 50 characters
- **Breaking changes**: Use "!" after type/scope AND include "BREAKING CHANGE:" section
- **Scope**: Optional, lowercase, short, aligned with code directory/module name
- **Body**: Use bullet points, explain WHY not just WHAT
- **Be specific**: Avoid vague titles like "update" or "fix stuff"

## Allowed Types

| Type     | Description                           |
| -------- | ------------------------------------- |
| feat     | New feature                           |
| fix      | Bug fix                               |
| chore    | Maintenance (e.g., tooling, deps)     |
| perf     | Performance improvements              |
| refactor | Code restructure (no behavior change) |
| docs     | Documentation changes                 |
| test     | Adding or refactoring tests           |
| style    | Code formatting (no logic change)     |
| build    | Changes that affect the build system  |
| ci       | Changes to CI configuration           |

## Workflow

1. **Review changes in parallel**:
   - Run `git status` (never use `-uall` flag)
   - Run `git diff` (staged and unstaged)
   - Run `git log -5 --oneline` to see recent commit style

2. **Draft commit message**:
   - Summarize the nature of changes (new feature, bug fix, etc.)
   - Ensure message accurately reflects changes and purpose
   - Focus on WHY rather than WHAT
   - Never commit files that likely contain secrets (.env, credentials.json, etc.)

3. **Stage and commit**:
   - Add relevant untracked files: `git add <files>`
   - Create commit with proper message format
   - Always use HEREDOC for commit messages:

   ```bash
   git commit -m "$(cat <<'EOF'
   feat(scope): descriptive title

   - Bullet point summary
   - Another change detail

   EOF
   )"
   ```

   - Verify with `git status` after commit

4. **If pre-commit hook fails**:
   - Fix the issue
   - Create a NEW commit (do not use `--amend`)
   - Never skip hooks unless explicitly requested

## Important Notes

- **Never push** unless user explicitly requests it
- **Never use interactive flags** (`-i`) as they require user input
- **Never use `--no-edit`** with git rebase (not a valid option)
- **No empty commits** - If nothing to commit, don't create one
- **Always use HEREDOC** for commit messages to ensure proper formatting
