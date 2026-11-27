import { View, div } from "gpui-kit";
import { h_flex, v_flex } from "gpui-base";
import { checking, closeOverlay, exec, focused, panes, props, spaces, theme } from "herdr";
import { projectKey, statusFromResult } from "./status.js";

const HAS_SPECLET = 'dir=$(pwd -P) || exit 2; while :; do if [ -e "$dir/speclet" ] || [ -L "$dir/speclet" ]; then exit 0; fi; parent=${dir%/*}; [ -n "$parent" ] || parent=/; [ "$parent" != "$dir" ] || exit 1; dir=$parent; done';

function currentProject() {
  const space = spaces().find((entry) => entry.active);
  const focus = focused();
  const pane = focus?.pane ? panes().find((entry) => entry.pane === focus.pane) : null;
  const tabs = new Set(space?.tabs.map((tab) => tab.id) ?? []);
  const firstPane = panes().find((entry) => tabs.has(entry.tab) && entry.cwd);
  return {
    cwd: pane?.cwd || focus?.cwd || firstPane?.cwd || space?.folder || null,
    on: space?.machine ?? null,
  };
}

function projectName(cwd) {
  return cwd?.split(/[\\/]/).filter(Boolean).at(-1) ?? cwd ?? "";
}

export default class SpecletStatus extends View {
  init(_props, cx) {
    this.project = currentProject();
    this.projectId = projectKey(this.project);
    this.requestId = 0;
    this.state = this.project.cwd ? { kind: "loading" } : { kind: "no-project" };
    if (!checking()) {
      this.refresh(cx);
      cx.timer.every(10000, (cx) => {
        if (!this.followProject(cx) && this.project.cwd) this.refresh(cx);
      });
    }
  }

  followProject(cx) {
    const next = currentProject();
    const nextId = projectKey(next);
    if (nextId === this.projectId) return false;
    this.project = next;
    this.projectId = nextId;
    this.state = next.cwd ? { kind: "loading" } : { kind: "no-project" };
    this.refresh(cx);
    return true;
  }

  refresh(cx) {
    const project = this.project;
    const id = projectKey(project);
    const requestId = ++this.requestId;
    if (!project.cwd) {
      this.state = { kind: "no-project" };
      cx.notify();
      return;
    }
    cx.spawn(async (cx) => {
      let next;
      const options = { cwd: project.cwd, on: project.on, timeout: 5000 };
      try {
        const directory = await exec("sh", ["-c", HAS_SPECLET], options);
        if (directory.code === 1) {
          next = statusFromResult({ initialized: false });
        } else if (directory.code !== 0) {
          next = { kind: "cli-error", message: directory.stderr.trim().split("\n")[0] || `檢查 speclet/ 結束碼 ${directory.code}` };
        } else {
          const result = await exec("speclet", ["change", "list", "--json"], options);
          next = statusFromResult({ initialized: true, result });
        }
      } catch (error) {
        next = statusFromResult({ initialized: true, error });
      }
      if (requestId !== this.requestId || id !== projectKey(currentProject())) return;
      this.state = next;
      cx.notify();
    });
  }

  render(cx) {
    this.followProject(cx);
    const t = theme();
    const overlay = props()?.overlay;
    const state = this.state;
    const message = {
      loading: "讀取 Speclet 任務狀態…",
      "no-project": "目前沒有可讀取的 Herdr 專案。",
      uninitialized: "這個專案尚未初始化 Speclet。請在專案根目錄執行 speclet init。",
      "missing-cli": "找不到 speclet 可執行檔。請依 Speclet README 安裝 CLI，並確認 Herdr 可從 PATH 執行 speclet。",
      "invalid-json": "Speclet CLI 回傳的 JSON 格式無效，請更新或檢查 CLI。",
      empty: "目前沒有進行中的 Speclet 變更。",
      "cli-error": state.message,
    }[state.kind];

    return v_flex()
      .size_full()
      .bg(t.popover)
      .text_color(t.foreground)
      .text_size(13)
      .p(10)
      .gap(8)
      .child(
        h_flex()
          .flex_none()
          .items_center()
          .gap(8)
          .child(v_flex().flex_1().child(div().font_semibold().child("Speclet 任務狀態")).child(
            div().text_size(11).text_color(t.muted).child(projectName(this.project.cwd)),
          ))
          .child(
            div()
              .px(8)
              .py(5)
              .rounded(5)
              .cursor_pointer()
              .hover((style) => style.bg(t.hover))
              .on_click(() => this.refresh(cx))
              .child("重新整理"),
          )
          .when(Boolean(overlay), (el) =>
            el.child(
              div()
                .px(8)
                .py(5)
                .rounded(5)
                .cursor_pointer()
                .hover((style) => style.bg(t.hover))
                .on_click(() => closeOverlay(overlay))
                .child("關閉"),
            ),
          ),
      )
      .child(
        v_flex()
          .id("speclet-status-body")
          .flex_1()
          .min_h(0)
          .overflow_y_scrollbar()
          .gap(4)
          .when(state.kind === "changes", (el) =>
            el.children(state.changes.map((change) =>
              h_flex()
                .id(`change-${change.name}`)
                .items_center()
                .gap(8)
                .px(8)
                .py(7)
                .rounded(5)
                .child(div().flex_1().min_w_0().truncate().child(change.name))
                .child(div().flex_none().text_color(t.muted).child(`${change.done}/${change.total}`)),
            )),
          )
          .when(state.kind !== "changes", (el) =>
            el.child(div().text_color(state.kind === "cli-error" ? t.blocked : t.muted).child(message ?? "讀取失敗。")),
          ),
      );
  }
}
