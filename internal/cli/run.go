package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/chiisen/speclet/internal/archive"
	"github.com/chiisen/speclet/internal/change"
	"github.com/chiisen/speclet/internal/project"
	"github.com/chiisen/speclet/internal/task"
)

// Run executes one Speclet command and returns its process exit code.
func Run(args []string, stdout, stderr io.Writer, cwd string) int {
	if len(args) == 0 {
		return fail(stderr, "missing command")
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		_, _ = io.WriteString(stdout, helpText)
		return 0
	}
	switch args[0] {
	case "init":
		if len(args) > 2 {
			return fail(stderr, "usage: speclet init [PATH]")
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
			return fail(stderr, "inspect speclet directory: %v", err)
		}
		if err := project.Init(target); err != nil {
			return fail(stderr, "speclet init: %v", err)
		}
		if alreadyInitialized {
			_, _ = fmt.Fprintln(stdout, "Speclet project already initialized.")
		} else {
			_, _ = fmt.Fprintln(stdout, "Initialized Speclet project.")
		}
		return 0
	case "change":
		return runChange(args[1:], stdout, stderr, cwd)
	case "task":
		return runTask(args[1:], stdout, stderr, cwd)
	case "archive":
		if len(args) != 2 {
			return fail(stderr, "usage: speclet archive <name>")
		}
		root, err := project.FindRoot(cwd)
		if err != nil {
			return fail(stderr, "%v", err)
		}
		if err := archive.Run(root, args[1]); err != nil {
			return fail(stderr, "archive change: %v", err)
		}
		_, _ = fmt.Fprintf(stdout, "Archived change %q.\n", args[1])
		return 0
	default:
		return fail(stderr, "unknown command: %s", args[0])
	}
}

const helpText = `Speclet - local-first specification workflow

Usage:
  speclet init [PATH]
  speclet change new <name>
  speclet change list
  speclet change show <name>
  speclet task done <change> <task-id>
  speclet task undone <change> <task-id>
  speclet archive <name>
  speclet help

Changes use Markdown proposal.md, tasks.md, and specs/<capability>/spec.md files.
Archive requires every task to be complete and every delta to be valid.
`

func runTask(args []string, stdout, stderr io.Writer, cwd string) int {
	if len(args) != 3 || (args[0] != "done" && args[0] != "undone") {
		return fail(stderr, "usage: speclet task <done|undone> <change> <task-id>")
	}
	root, err := project.FindRoot(cwd)
	if err != nil {
		return fail(stderr, "%v", err)
	}
	release, err := project.AcquireWriteLock(root)
	if err != nil {
		return fail(stderr, "%v", err)
	}
	detail, err := change.Read(root, args[1])
	if err != nil {
		_ = release()
		return fail(stderr, "find change: %v", err)
	}
	done := args[0] == "done"
	if err := task.SetDone(detail.TasksPath, args[2], done); err != nil {
		_ = release()
		return fail(stderr, "update task: %v", err)
	}
	if err := release(); err != nil {
		return fail(stderr, "%v", err)
	}
	state := "undone"
	if done {
		state = "done"
	}
	_, _ = fmt.Fprintf(stdout, "Marked task %s %s for change %q.\n", args[2], state, detail.Name)
	return 0
}

func runChange(args []string, stdout, stderr io.Writer, cwd string) int {
	if len(args) == 0 {
		return fail(stderr, "usage: speclet change <new|list|show>")
	}
	root, err := project.FindRoot(cwd)
	if err != nil {
		return fail(stderr, "%v", err)
	}
	switch args[0] {
	case "new":
		if len(args) != 2 {
			return fail(stderr, "usage: speclet change new <name>")
		}
		if err := change.Create(root, args[1]); err != nil {
			return fail(stderr, "create change: %v", err)
		}
		_, _ = fmt.Fprintf(stdout, "Created change %q.\n", args[1])
		return 0
	case "list":
		if len(args) != 1 {
			return fail(stderr, "usage: speclet change list")
		}
		summaries, err := change.List(root)
		if err != nil {
			return fail(stderr, "list changes: %v", err)
		}
		if len(summaries) == 0 {
			_, _ = fmt.Fprintln(stdout, "No active changes.")
			return 0
		}
		for _, summary := range summaries {
			_, _ = fmt.Fprintf(stdout, "%s (%d/%d tasks done)\n", summary.Name, summary.Done, summary.Total)
		}
		return 0
	case "show":
		if len(args) != 2 {
			return fail(stderr, "usage: speclet change show <name>")
		}
		detail, err := change.Read(root, args[1])
		if err != nil {
			return fail(stderr, "show change: %v", err)
		}
		_, _ = fmt.Fprintf(stdout, "Change: %s\nProposal: %s\nTasks: %s\n", detail.Name, detail.ProposalPath, detail.TasksPath)
		if len(detail.DeltaPaths) == 0 {
			_, _ = fmt.Fprintln(stdout, "Delta specs: none")
		} else {
			_, _ = fmt.Fprintln(stdout, "Delta specs:")
			for _, path := range detail.DeltaPaths {
				_, _ = fmt.Fprintf(stdout, "  %s\n", path)
			}
		}
		return 0
	default:
		return fail(stderr, "unknown change command: %s", args[0])
	}
}

func fail(stderr io.Writer, format string, args ...any) int {
	_, _ = fmt.Fprintf(stderr, format+"\n", args...)
	return 1
}
