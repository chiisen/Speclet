# Speclet Herdr 任務狀態

這是 Speclet 的選用 Herdr mod。它會在目前 tab 的右側 pane 顯示目前專案的進行中變更及已完成／總任務數。Speclet CLI 本身不依賴 Herdr。

## 安裝

先依 repo 根目錄 README 安裝可執行的 `speclet` CLI，再安裝 mod。Herdr 主機執行 `speclet` 時必須能從 `PATH` 找到 CLI；遠端 space 需在遠端機器安裝 CLI。

macOS／Linux 將 repo 路徑代換為此 repo 的位置：

~~~sh
HERDR_APP_HOME="$HOME/.config/herdr-app"
mkdir -p "$HERDR_APP_HOME/mods"
cp -R /path/to/Speclet/integrations/herdr/mods/speclet-status "$HERDR_APP_HOME/mods/"
~~~

Windows PowerShell 範例：

~~~powershell
$repo = 'D:\path\to\Speclet'
$herdrHome = if ($env:HERDR_APP_HOME) { $env:HERDR_APP_HOME } else { Join-Path $HOME '.config\herdr-app' }
$mods = Join-Path $herdrHome 'mods'
New-Item -ItemType Directory -Force -Path $mods | Out-Null
Copy-Item -Recurse (Join-Path $repo 'integrations\herdr\mods\speclet-status') $mods
~~~

## 開啟右側面板

- Windows／Linux：在 Herdr 按 Ctrl+Shift+P。
- macOS：在 Herdr 按 Command+Shift+P。
- 搜尋並選取「Speclet：任務狀態」。

也可以從目前焦點 pane 右側開啟：

~~~sh
herdr-app do pane.split '{"direction":"right","mod":"speclet-status"}'
~~~

面板留在目前 tab，可使用 Herdr pane 控制調整、關閉或再次開啟。分割比例由 Herdr 決定，可拖曳分隔線縮窄；顯示任務狀態通常不需要一半寬度。它會跟隨目前 Herdr 專案；切換專案時清除舊資料並重新讀取。每 10 秒自動更新，也可以按「重新整理」。

## 限制

- mod 只執行 `speclet change list --json` 和檢查 `speclet/` 是否存在；不會寫入 Speclet 資料、勾選任務、派工或呼叫 agent。
- 專案未初始化、CLI 不存在、CLI 失敗和輸出不是有效 JSON 會顯示不同提示，不會把錯誤當成空清單。
- Herdr 桌面 pane 需在 Herdr 中實際開啟才完成畫面驗收；純 Node 測試不代表桌面畫面已驗證。

## 測試

~~~sh
cd integrations/herdr
node --test
~~~
