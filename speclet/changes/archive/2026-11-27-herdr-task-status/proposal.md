# Change: herdr-task-status

## Why

Speclet tracks feature work in local Markdown files, but users currently need to leave Herdr and run the CLI to inspect task progress. A small optional Herdr mod can surface this read-only status without adding a Herdr dependency to Speclet itself.

## What Changes

- Add `speclet change list --json` with stable lower-camel-case fields and preserve the existing text and `--names-only` formats.
- Add an optional Herdr popup mod that invokes the current project's Speclet CLI and displays active changes and task counts.
- Document installation, opening the popup, CLI requirements, and the read-only boundary.
