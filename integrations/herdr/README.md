# Speclet Herdr 任務狀態

這是 Speclet 的選用 Herdr mod。它會讀取目前 Herdr 專案的進行中變更，並顯示各變更的已完成／總任務數。Speclet CLI 本身不依賴 Herdr。

## 安裝

先依 repo 根目錄 README 安裝可執行的 `speclet` CLI。接著在 Herdr 使用者資料目錄安裝 mod；將 `/path/to/Speclet` 換成 repo 的位置：

```sh
HERDR_APP_HOME="${HERDR_APP_HOME:-$HOME/.config/herdr-app}"
mkdir -p "$HERDR_APP_HOME/mods"
cp -R /path/to/Speclet/integrations/herdr/mods/speclet-status "$HERDR_APP_HOME/mods/"
```

Herdr 主機執行 `speclet` 時必須能從 `PATH` 找到 CLI。本機專案使用本機 CLI；遠端 space 則需在該遠端機器安裝 CLI。

## 開啟面板

- 在 Herdr 按 `⌘⇧P`，搜尋並選取「Speclet：任務狀態」。
- 或執行 `herdr-app do mod.open '{"name":"speclet-status"}'`。

面板會跟隨目前焦點終端機所在的專案，切換專案時清除舊資料並重新讀取。每 10 秒自動更新，也可以按「重新整理」。

## 限制

- mod 只執行 `speclet change list --json` 和檢查 `speclet/` 是否存在；不會寫入 Speclet 資料、勾選任務、派工或呼叫 agent。
- 專案未初始化、CLI 不存在、CLI 失敗和輸出不是有效 JSON 會顯示不同提示，不會把錯誤當成空清單。
- Herdr 桌面面板需在 Herdr 中實際開啟才完成畫面驗收；純 Node 測試不代表桌面畫面已驗證。

## 測試

```sh
cd integrations/herdr
node --test
```
