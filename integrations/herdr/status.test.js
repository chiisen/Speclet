import assert from "node:assert/strict";
import test from "node:test";

const model = await import("./mods/speclet-status/status.js").catch(() => ({}));

test("status model is available", () => {
  assert.equal(typeof model.statusFromResult, "function");
});

test("uninitialized project is distinct from an empty change list", () => {
  assert.deepEqual(model.statusFromResult({ initialized: false }), { kind: "uninitialized" });
  assert.deepEqual(model.statusFromResult({ initialized: true, result: { code: 0, stdout: "[]\n" } }), { kind: "empty" });
});

test("missing Speclet executable gives an actionable state", () => {
  const error = Object.assign(new Error("spawn speclet ENOENT"), { code: "ENOENT" });
  assert.deepEqual(model.statusFromResult({ initialized: true, error }), { kind: "missing-cli" });
});

test("CLI failure and invalid JSON are not treated as an empty list", () => {
  assert.deepEqual(model.statusFromResult({
    initialized: true,
    result: { code: 2, stdout: "[]\n", stderr: "invalid project data" },
  }), { kind: "cli-error", message: "invalid project data" });
  assert.deepEqual(model.statusFromResult({
    initialized: true,
    result: { code: 0, stdout: "not json", stderr: "" },
  }), { kind: "invalid-json" });
  assert.deepEqual(model.statusFromResult({
    initialized: true,
    result: { code: 0, stdout: '[{"name":"broken","done":"2","total":3}]', stderr: "" },
  }), { kind: "invalid-json" });
});

test("valid JSON preserves all change names and task counts", () => {
  assert.deepEqual(model.statusFromResult({
    initialized: true,
    result: {
      code: 0,
      stdout: '[{"name":"herdr-task-status","done":2,"total":4},{"name":"付款 checkout","done":0,"total":3}]\n',
      stderr: "",
    },
  }), {
    kind: "changes",
    changes: [
      { name: "herdr-task-status", done: 2, total: 4 },
      { name: "付款 checkout", done: 0, total: 3 },
    ],
  });
});

test("project key includes machine and working directory", () => {
  assert.notEqual(model.projectKey({ cwd: "/work/a", on: null }), model.projectKey({ cwd: "/work/b", on: null }));
  assert.notEqual(model.projectKey({ cwd: "/work/a", on: "remote" }), model.projectKey({ cwd: "/work/a", on: null }));
});
