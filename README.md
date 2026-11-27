# Speclet

Speclet 是一個本機優先的規格驅動開發 CLI，目標是在單一專案中，以容易閱讀的檔案管理規格、變更提案與任務清單。

## 專案狀態

MVP 已採 Go 實作，提供初始化、變更管理、任務追蹤和安全封存功能。

## 安裝

需要 Go 1.27 或更新版本。在 repo 根目錄執行：

```sh
go install -buildvcs=true ./cmd/speclet
```

`go install` 會把執行檔安裝到 Go 的 bin 目錄。預設位置是 `$(go env GOPATH)/bin`；若有設定 `GOBIN`，則使用 `GOBIN` 指定的位置。安裝後確認命令可用：

```sh
command -v speclet
speclet help
speclet --version
```

如果 zsh 顯示 `command not found: speclet`，先確認 `go install -buildvcs=true ./cmd/speclet` 已成功執行。若執行檔已安裝，但 Go 的 bin 目錄尚未加入 `PATH`，可先在目前終端機暫時加入預設目錄：

```sh
export PATH="$(go env GOPATH)/bin:$PATH"
```

確認可執行後，將這行加入 `~/.zshrc`，並執行 `source ~/.zshrc`，之後開啟的新終端機也會套用設定。如果有設定 `GOBIN`，請改為將該目錄加入 `PATH`。

也可以建置獨立執行檔：

```sh
go build -buildvcs=true -o /tmp/speclet ./cmd/speclet
```

### 版本資訊與正式建置

`speclet --version` 及 help 會顯示執行檔版本。一般 Go 建置以 `dev+` 和 Git commit 短 SHA 標示；沒有 VCS revision 時顯示 `dev`。請以 `-buildvcs=true` 建置，確保 Go 將 revision 寫入 build info。

正式版使用 `vMAJOR.MINOR.PATCH` Git tag 作為版本，例如快速演進期間的 `v0.1.0`。在標籤所指向的 commit 建置；沒有精確 tag 時，以下命令會停止，不會產生未標記的正式版本。

Windows PowerShell：

```powershell
$version = git describe --tags --exact-match HEAD
if ($LASTEXITCODE -ne 0) { throw '目前 commit 沒有正式版本 tag。' }
go build -buildvcs=true -ldflags "-X github.com/chiisen/speclet/internal/version.tag=$version" -o speclet.exe ./cmd/speclet
```

macOS／Linux：

```sh
version=$(git describe --tags --exact-match HEAD) || exit 1
go build -buildvcs=true -ldflags "-X github.com/chiisen/speclet/internal/version.tag=$version" -o speclet ./cmd/speclet
```

## 快速開始

第一次使用可以跟著[第一個 Speclet 變更](docs/tutorials/first-change.md)操作。
AI coding agent 的功能開發流程另見[使用 Speclet 開發功能](docs/tutorials/ai-feature-workflow.md)。
只想看任務狀態範例，請參考[用 Speclet 管理任務狀態](docs/tutorials/task-status.md)。

在要管理規格的專案根目錄初始化資料並建立變更：

```sh
cd /path/to/your-project
speclet init
speclet change new payment-checkout
```

Speclet 建立以下結構：

```text
speclet/
  config.yaml
  specs/<capability>/spec.md
  changes/<change-name>/
    proposal.md
    tasks.md
    specs/<capability>/spec.md
  changes/archive/<YYYY-MM-DD>-<change-name>/
```

編輯變更目錄裡的 `proposal.md`、`tasks.md` 和規格差異。任務使用平面 checkbox，並為每項工作指定唯一 ID：

```markdown
- [ ] T001 新增付款流程
- [ ] T002 更新正式規格
```

規格差異使用 `ADDED`、`MODIFIED`、`REMOVED` 區段。例如新增需求：

```markdown
## ADDED

### Requirement: Customer can start a payment

#### Scenario: Valid payment details
- **WHEN** the customer submits valid details
- **THEN** a payment is created
```

完成工作並封存：

```sh
speclet change list
speclet change list --json
speclet change show payment-checkout
speclet task done payment-checkout T001
speclet task done payment-checkout T002
speclet archive payment-checkout
```

`archive` 會先確認所有任務已完成、需求名稱符合新增／修改／移除規則，且目標規格路徑有效；驗證通過後才更新正式規格，並將變更移到日期封存目錄。失敗時會保留進行中的變更和正式規格。

同一專案的任務狀態更新與封存會互斥執行。若程序遭強制終止而留下 `speclet/.operation.lock/`，先確認沒有其他 Speclet 寫入程序，再移除該鎖目錄後重試。

## 命令

```text
speclet init [PATH]
speclet change new <name>
speclet change list
speclet change show <name>
speclet task done <change> <task-id>
speclet task undone <change> <task-id>
speclet archive <name>
speclet help
```

`speclet change list` 會列出進行中的變更和任務進度。加上 `--names-only` 後，每行只輸出一個完整名稱，方便腳本讀取；加上 `--json` 後則輸出穩定的 JSON 摘要格式。兩個旗標不能同時使用。預設輸出與 `--names-only` 沒有變更；沒有進行中的變更時，前者顯示提示、後者輸出空白，JSON 模式輸出 `[]`。變更名稱含空格或 Unicode 時會保留。

JSON 每筆使用小駝峰欄位 `name`、`done` 和 `total`：

```json
[{"name":"payment-checkout","done":2,"total":4}]
```

### Herdr 選用整合

`integrations/herdr/` 提供只讀的 Herdr 任務狀態 pane，顯示目前專案的進行中變更和任務完成數。這是獨立選用的 mod，不是 Speclet CLI 的執行期依賴。

在 Herdr 主機安裝 mod。macOS／Linux 將 repo 路徑代換為此 repo 的位置：

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

在 Windows／Linux 按 Ctrl+Shift+P，或在 macOS 按 Command+Shift+P 開啟指令面板，搜尋「Speclet：任務狀態」。也可以執行：

~~~sh
herdr-app do pane.split '{"direction":"right","mod":"speclet-status"}'
~~~

面板會出現在目前 tab 焦點 pane 的右側，並隨目前 Herdr 專案切換而重新讀取狀態。預設分割比例由 Herdr 決定；可拖曳分隔線調整，狀態面板通常不需要一半寬度。每 10 秒自動更新，也可以按「重新整理」。Herdr 執行環境必須能從 `PATH` 找到 `speclet` CLI；完整安裝與限制見 [`integrations/herdr/README.md`](integrations/herdr/README.md)。

### 語系設定

Speclet 預設以英文顯示。可先查詢支援語系，再設定繁體中文；設定會保存在目前電腦的使用者設定目錄，套用到之後執行的所有專案：

```sh
speclet config language list
speclet config language show
speclet config language set zh-TW
```

目前支援 `en`（English）和 `zh-TW`（繁體中文）。要切回英文時執行 `speclet config language set en`。未列出的語系代碼會被拒絕，不會改動既有設定。

Speclet 不會操作 Git commit、branch 或 push，也不會呼叫 AI 模型。AI coding agent 可透過這些檔案和命令參與工作流程。
