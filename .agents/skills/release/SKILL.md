---
name: release
description: Run the OneTiny release-please release workflow. Use when the user invokes $release, $release pub, asks to release OneTiny, publish a release, merge a release-please PR, or finish the release process. Supports default draft-only release and pub mode that publishes the draft after assets are uploaded.
---

# OneTiny Release Skill

Release OneTiny through release-please. Do not manually create tags or run `gh release create`.

## Modes

- `$release`: Merge the existing release-please PR, wait for workflows, verify the draft release and assets, then stop with the release still draft.
- `$release pub`: Do the same, then publish the draft release after assets are present.

Merging feature PRs into `main` is release preparation, not part of this release workflow. Do not merge feature PRs unless the user explicitly asks for that separately.

## Preconditions

- Work in the OneTiny repo, or pass `--repo tcp404/OneTiny` to GitHub CLI commands.
- Use `rtk` for shell commands in this project.
- A release-please PR or an existing draft release must already exist. If neither exists, report that there is no release to perform instead of creating tags or releases manually.
- If the version or changelog looks wrong, stop and report the mismatch before merging.

## State Handling

- If a release-please PR is open, use it as the source of the target version and tag.
- If no release-please PR is open but a draft release already exists:
  - For `$release pub`, verify its assets and publish it.
  - For `$release`, report the existing draft and do not run a new release.
- If multiple draft releases exist, ask the user which tag to publish.
- If the target release is already public, report that state and do not rerun release commands.

## Workflow

1. Find the open release-please PR:

   ```bash
   rtk gh pr list --state open --json number,title,headRefName,baseRefName,author,isDraft,mergeable,updatedAt,url --repo tcp404/OneTiny
   ```

   Select the PR whose head branch starts with `release-please--branches--main` and whose base is `main`.

   If none exists, inspect existing releases before stopping:

   ```bash
   rtk gh release list --limit 10 --repo tcp404/OneTiny
   ```

2. Inspect the release PR:

   ```bash
   rtk gh pr view <PR_NUMBER> --json number,title,state,body,headRefName,baseRefName,mergeable,url --repo tcp404/OneTiny
   ```

   Confirm:
   - title version matches the changelog body,
   - body contains the intended release notes,
   - PR is open, not draft, and mergeable.

3. Squash merge the release PR:

   ```bash
   rtk gh pr merge <PR_NUMBER> --squash --delete-branch --repo tcp404/OneTiny
   ```

4. Watch the release workflow started by the merge:

   ```bash
   rtk gh run list --limit 10 --repo tcp404/OneTiny
   ```

   Wait until the new `Release Please` run completes successfully. Prefer low-noise polling over long repeated logs:

   ```bash
   rtk gh run view <RUN_ID> --json status,conclusion,jobs --repo tcp404/OneTiny
   ```

   If it fails, inspect the failed job logs and report the failure. Do not publish.

5. Verify the GitHub Release by tag:

   ```bash
   rtk gh release view <TAG> --json tagName,name,isDraft,isPrerelease,publishedAt,assets,url --repo tcp404/OneTiny
   ```

   Expected assets:
   - `onetiny-checksums.txt`
   - `onetiny-cli-darwin-arm64.zip`
   - `onetiny-cli-darwin-x64.zip`
   - `onetiny-cli-linux-x64.zip`
   - `onetiny-cli-windows-x64.zip`
   - `onetiny-gui-darwin-arm64.zip`
   - `onetiny-gui-windows-x64.zip`

6. For `$release`, stop here and report the draft release URL, tag, workflow status, and asset names.

7. For `$release pub`, publish the draft only after all expected assets are uploaded:

   ```bash
   rtk gh release edit <TAG> --draft=false --repo tcp404/OneTiny
   ```

   Verify `isDraft: false` afterward and report the public release URL.

## Guardrails

- Never use `git tag`, `gh release create`, or manual changelog edits during normal release.
- Never publish a draft with missing assets.
- Never publish if the release workflow failed or is still running.
- If the release is already public, report that state and do not re-run release commands.
- If GitHub keeps the final upload job queued, wait and re-check; do not publish until assets exist.
