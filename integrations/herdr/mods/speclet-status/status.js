export function projectKey(project) {
  return JSON.stringify([project?.on ?? null, project?.cwd ?? null]);
}

export function statusFromResult({ initialized, result, error }) {
  if (!initialized) return { kind: "uninitialized" };
  if (error) {
    const message = String(error?.message ?? error);
    if (error?.code === "ENOENT" || /\bENOENT\b|command not found|executable not found|program not installed/i.test(message)) {
      return { kind: "missing-cli" };
    }
    return { kind: "cli-error", message: message.split("\n")[0] || "無法啟動 Speclet CLI" };
  }
  if (!result || result.code !== 0) {
    const message = result?.stderr?.trim().split("\n")[0]
      || `speclet change list --json 結束碼 ${result?.code ?? "unknown"}`;
    return { kind: "cli-error", message };
  }

  let changes;
  try {
    changes = JSON.parse(result.stdout);
  } catch {
    return { kind: "invalid-json" };
  }
  if (!Array.isArray(changes) || !changes.every((change) =>
    change !== null
      && typeof change === "object"
      && typeof change.name === "string"
      && change.name.trim() !== ""
      && Number.isSafeInteger(change.done)
      && change.done >= 0
      && Number.isSafeInteger(change.total)
      && change.total >= change.done,
  )) {
    return { kind: "invalid-json" };
  }
  return changes.length === 0 ? { kind: "empty" } : { kind: "changes", changes };
}
