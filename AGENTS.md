# Speclet repository agent guide

## Use Speclet for feature work

When implementing a new user-visible feature or changing existing behavior in this repository, track the work with this repository's Speclet CLI:

1. Read the repository README for installation and PATH setup. Confirm the CLI is available with `command -v speclet`.
2. If the repository does not yet contain a `speclet/` data directory, initialize it with `speclet init` from the repository root. If initialization fails or the existing layout is invalid, stop and report the error instead of replacing project data.
3. Create a change with `speclet change new <name>`. Add a concise rationale to `proposal.md`, acceptance tasks to `tasks.md`, and a delta under `specs/<capability>/spec.md` for each affected capability.
4. Implement the change and keep task state in sync with `speclet task done` or `speclet task undone`. Add or update tests for behavior changes and run the relevant checks.
5. Before finishing, review the change with `speclet change show <name>` and `speclet change list`. Archive only after every task is complete and the delta describes the delivered behavior: `speclet archive <name>`.
6. Report the Speclet change name, archive result, and any checks that could not be run in the final response. Speclet manages local specification files; handle Git commits, branches, and pushes separately under the user's instructions.

Documentation-only edits and mechanical maintenance that do not change product behavior do not need a Speclet change.

## Handle defects found during Speclet work

Use the current task's agreed acceptance criteria and scope to decide what to do.

- **Within scope and affects acceptance.** Fix the defect in the current change, rerun the original failing case, and report the result with the task. A separate issue is unnecessary.
- **Blocks acceptance and requires a scope expansion.** Tell the user what is blocked and what additional changes are needed. Continue any unaffected work while the user decides whether to expand the scope.
- **Confirmed and outside the current task's scope, with acceptance unaffected.** Check for an existing issue in `chiisen/Speclet`. If one covers the same defect, report its link. Otherwise, create a GitHub issue with the affected version, minimal reproduction steps, expected and observed results, and impact. Remove secrets and personal data from the report. Continue the original task, report the issue link, and leave the repair schedule to the user. If you cannot check for duplicates or create the issue, include a ready-to-file draft and the reason in the final report.
- **Unconfirmed observation or improvement idea.** Include it in the task report when it has a concrete impact. Open bug issues for confirmed defects only.

When a tracked defect is repaired later, record the result of rerunning the original case. For a verified fix in a pull request targeting the default branch, use `Fixes #<issue-number>` to close the issue on merge.
