## ADDED

### Requirement: Machine-readable active change summaries

The CLI SHALL provide `speclet change list --json` as a stable machine-readable view of active changes. Each array item SHALL use the lower-camel-case fields `name`, `done`, and `total`; task counts SHALL come from the existing task counter. An empty project SHALL emit `[]`. JSON SHALL be the only stdout output for this mode, and invalid or conflicting arguments SHALL report an error on stderr and exit nonzero. The default human-readable output and `--names-only` behavior SHALL remain unchanged.

#### Scenario: List active changes as JSON

- **WHEN** `speclet change list --json` runs in an initialized project with active changes
- **THEN** stdout contains a JSON array of active change summaries with accurate task counts

#### Scenario: List an empty project as JSON

- **WHEN** `speclet change list --json` runs with no active changes
- **THEN** stdout is `[]` followed by a newline

#### Scenario: Reject incompatible or unsupported list arguments

- **WHEN** `--json` is combined with `--names-only` or an unsupported list argument is supplied
- **THEN** the CLI reports usage on stderr, emits no stdout, and exits nonzero

### Requirement: Optional read-only Herdr task status popup

The repository SHALL provide an optional Herdr mod under `integrations/herdr/`. The mod SHALL determine the current Herdr project's directory and machine, invoke the `speclet` executable with an argument array and that directory as its working directory, and display every active change name and its done/total task count in a popup. It SHALL only read project state. It SHALL distinguish an uninitialized project, a missing CLI, CLI failure, invalid JSON, and an initialized project with no active changes. A project switch SHALL clear the previous project's state and load the new project's status.

#### Scenario: Show active task progress

- **WHEN** the current project contains active Speclet changes and the Speclet CLI succeeds
- **THEN** the popup shows every change's name and completed/total task count

#### Scenario: Show empty, uninitialized, and error states

- **WHEN** the project is empty, uninitialized, missing the CLI, or the CLI fails or emits invalid JSON
- **THEN** the popup shows a distinct readable state and does not present a CLI error as an empty list

#### Scenario: Switch projects

- **WHEN** the active Herdr project changes
- **THEN** the popup discards prior project data and refreshes from the new project
