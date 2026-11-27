package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/chiisen/speclet/internal/archive"
	"github.com/chiisen/speclet/internal/change"
	"github.com/chiisen/speclet/internal/i18n"
	"github.com/chiisen/speclet/internal/project"
	"github.com/chiisen/speclet/internal/settings"
	"github.com/chiisen/speclet/internal/task"
)

const englishHelp = `Speclet - local-first specification workflow

Usage:
  speclet init [PATH]
  speclet change new <name>
  speclet change list [--names-only | --json]
  speclet change show <name>
  speclet task done <change> <task-id>
  speclet task undone <change> <task-id>
  speclet archive <name>
  speclet config language list
  speclet config language show
  speclet config language set <code>
  speclet help

Use --names-only to print one active change name per line without task counts.
Use --json to print active change summaries as a JSON array.
Changes use Markdown proposal.md, tasks.md, and specs/<capability>/spec.md files.
Archive requires every task to be complete and every delta to be valid.
`

func Run(args []string, stdout, stderr io.Writer, cwd string) int {
	locale, configured, settingErr := settings.Language()
	if settingErr != nil {
		locale = i18n.English
		configured = false
	}
	if len(args) == 0 {
		return fail(stderr, locale, "missing command")
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		_, _ = fmt.Fprintln(stdout, i18n.Text(locale, englishHelp))
		return 0
	}
	if args[0] == "config" {
		return runConfig(args[1:], stdout, stderr, locale, configured, settingErr)
	}
	if settingErr != nil {
		return failError(stderr, i18n.English, settingErr)
	}
	switch args[0] {
	case "init":
		if len(args) > 2 {
			return fail(stderr, locale, "usage: speclet init [PATH]")
		}
		target := cwd
		if len(args) == 2 {
			target = args[1]
			if !filepath.IsAbs(target) {
				target = filepath.Join(cwd, target)
			}
		}
		alreadyInitialized := false
		if _, err := os.Lstat(filepath.Join(target, "speclet")); err == nil {
			alreadyInitialized = true
		} else if !os.IsNotExist(err) {
			return failError(stderr, locale, i18n.Errorf("inspect speclet directory: %w", err))
		}
		if err := project.Init(target); err != nil {
			return failError(stderr, locale, i18n.Errorf("speclet init: %w", err))
		}
		if alreadyInitialized {
			_, _ = fmt.Fprintln(stdout, i18n.Text(locale, "Speclet project already initialized."))
		} else {
			_, _ = fmt.Fprintln(stdout, i18n.Text(locale, "Initialized Speclet project."))
		}
		return 0
	case "change":
		return runChange(args[1:], stdout, stderr, cwd, locale)
	case "task":
		return runTask(args[1:], stdout, stderr, cwd, locale)
	case "archive":
		if len(args) != 2 {
			return fail(stderr, locale, "usage: speclet archive <name>")
		}
		root, err := project.FindRoot(cwd)
		if err != nil {
			return failError(stderr, locale, err)
		}
		if err := archive.Run(root, args[1]); err != nil {
			return failError(stderr, locale, i18n.Errorf("archive change: %w", err))
		}
		_, _ = fmt.Fprintf(stdout, i18n.Text(locale, "Archived change %q.")+"\n", args[1])
		return 0
	default:
		return fail(stderr, locale, "unknown command: %s", args[0])
	}
}

func runConfig(args []string, stdout, stderr io.Writer, locale string, configured bool, settingErr error) int {
	if len(args) < 2 || args[0] != "language" {
		return fail(stderr, locale, "usage: speclet config language <list|show|set> [code]")
	}
	switch args[1] {
	case "list":
		if len(args) != 2 {
			return fail(stderr, locale, "usage: speclet config language <list|show|set> [code]")
		}
		_, _ = fmt.Fprintln(stdout, i18n.Text(locale, "Supported languages:"))
		for _, language := range i18n.SupportedLanguages() {
			_, _ = fmt.Fprintf(stdout, "%s\t%s\n", language.Code, i18n.Name(language.Code, locale))
		}
		return 0
	case "show":
		if len(args) != 2 {
			return fail(stderr, locale, "usage: speclet config language <list|show|set> [code]")
		}
		if settingErr != nil {
			return failError(stderr, i18n.English, settingErr)
		}
		name := i18n.Name(locale, locale)
		if configured {
			_, _ = fmt.Fprintf(stdout, i18n.Text(locale, "Current language: %s (%s).")+"\n", name, locale)
		} else {
			_, _ = fmt.Fprintf(stdout, i18n.Text(locale, "Current language: %s (%s, default).")+"\n", name, locale)
		}
		return 0
	case "set":
		if len(args) != 3 {
			return fail(stderr, locale, "usage: speclet config language set <code>")
		}
		if err := settings.SetLanguage(args[2]); err != nil {
			return failError(stderr, locale, err)
		}
		_, _ = fmt.Fprintf(stdout, i18n.Text(args[2], "Language preference saved: %s (%s).")+"\n", i18n.Name(args[2], args[2]), args[2])
		return 0
	default:
		return fail(stderr, locale, "usage: speclet config language <list|show|set> [code]")
	}
}

func runTask(args []string, stdout, stderr io.Writer, cwd, locale string) int {
	if len(args) != 3 || (args[0] != "done" && args[0] != "undone") {
		return fail(stderr, locale, "usage: speclet task <done|undone> <change> <task-id>")
	}
	root, err := project.FindRoot(cwd)
	if err != nil {
		return failError(stderr, locale, err)
	}
	release, err := project.AcquireWriteLock(root)
	if err != nil {
		return failError(stderr, locale, err)
	}
	detail, err := change.Read(root, args[1])
	if err != nil {
		_ = release()
		return failError(stderr, locale, i18n.Errorf("find change: %w", err))
	}
	done := args[0] == "done"
	if err := task.SetDone(detail.TasksPath, args[2], done); err != nil {
		_ = release()
		return failError(stderr, locale, i18n.Errorf("update task: %w", err))
	}
	if err := release(); err != nil {
		return failError(stderr, locale, err)
	}
	state := i18n.Text(locale, "undone")
	if done {
		state = i18n.Text(locale, "done")
	}
	_, _ = fmt.Fprintf(stdout, i18n.Text(locale, "Marked task %s %s for change %q.")+"\n", args[2], state, detail.Name)
	return 0
}

func runChange(args []string, stdout, stderr io.Writer, cwd, locale string) int {
	if len(args) == 0 {
		return fail(stderr, locale, "usage: speclet change <new|list|show>")
	}
	root, err := project.FindRoot(cwd)
	if err != nil {
		return failError(stderr, locale, err)
	}
	switch args[0] {
	case "new":
		if len(args) != 2 {
			return fail(stderr, locale, "usage: speclet change new <name>")
		}
		if err := change.Create(root, args[1]); err != nil {
			return failError(stderr, locale, i18n.Errorf("create change: %w", err))
		}
		_, _ = fmt.Fprintf(stdout, i18n.Text(locale, "Created change %q.")+"\n", args[1])
		return 0
	case "list":
		namesOnly, jsonOutput := false, false
		for _, arg := range args[1:] {
			switch arg {
			case "--names-only":
				namesOnly = true
			case "--json":
				jsonOutput = true
			default:
				return fail(stderr, locale, "usage: speclet change list [--names-only | --json]")
			}
		}
		if len(args) > 2 || (namesOnly && jsonOutput) {
			return fail(stderr, locale, "use either --json or --names-only, not both; usage: speclet change list [--names-only | --json]")
		}
		summaries, err := change.List(root)
		if err != nil {
			return failError(stderr, locale, i18n.Errorf("list changes: %w", err))
		}
		if jsonOutput {
			if err := json.NewEncoder(stdout).Encode(summaries); err != nil {
				return failError(stderr, locale, i18n.Errorf("write change list JSON: %w", err))
			}
			return 0
		}
		if len(summaries) == 0 && !namesOnly {
			_, _ = fmt.Fprintln(stdout, i18n.Text(locale, "No active changes."))
			return 0
		}
		for _, summary := range summaries {
			if namesOnly {
				_, _ = fmt.Fprintln(stdout, summary.Name)
			} else {
				_, _ = fmt.Fprintf(stdout, i18n.Text(locale, "%s (%d/%d tasks done)")+"\n", summary.Name, summary.Done, summary.Total)
			}
		}
		return 0
	case "show":
		if len(args) != 2 {
			return fail(stderr, locale, "usage: speclet change show <name>")
		}
		detail, err := change.Read(root, args[1])
		if err != nil {
			return failError(stderr, locale, i18n.Errorf("show change: %w", err))
		}
		_, _ = fmt.Fprintf(stdout, i18n.Text(locale, "Change: %s\nProposal: %s\nTasks: %s\n"), detail.Name, detail.ProposalPath, detail.TasksPath)
		if len(detail.DeltaPaths) == 0 {
			_, _ = fmt.Fprintln(stdout, i18n.Text(locale, "Delta specs: none"))
		} else {
			_, _ = fmt.Fprintln(stdout, i18n.Text(locale, "Delta specs:"))
			for _, path := range detail.DeltaPaths {
				_, _ = fmt.Fprintf(stdout, "  %s\n", path)
			}
		}
		return 0
	default:
		return fail(stderr, locale, "unknown change command: %s", args[0])
	}
}

func fail(stderr io.Writer, locale, key string, args ...any) int {
	_, _ = fmt.Fprintln(stderr, i18n.Text(locale, key, args...))
	return 1
}

func failError(stderr io.Writer, locale string, err error) int {
	_, _ = fmt.Fprintln(stderr, i18n.ErrorText(locale, err))
	return 1
}
