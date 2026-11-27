package version

import "runtime/debug"

// tag is set to the exact Git tag for release builds with -ldflags.
var tag = "dev"

func String() string {
	revision := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				revision = setting.Value
				break
			}
		}
	}
	return format(tag, revision)
}

func format(tag, revision string) string {
	if tag != "" && tag != "dev" {
		return tag
	}
	if len(revision) < 7 {
		return "dev"
	}
	return "dev+" + revision[:7]
}
