# Speclet 本機 CLI MVP 實作計畫

> **給 agentic workers：** 實作時必須使用 `superpowers:subagent-driven-development` 或 `superpowers:executing-plans` 子技能，逐項完成本計畫。每個步驟使用 checkbox 追蹤。

**目標：** 建立一個 Go 本機 CLI，讓開發者初始化 `speclet/` 資料、建立和追蹤變更，並安全地把有效規格差異封存到正式規格。

**架構：** `cmd/speclet` 是唯一執行入口；`internal/cli` 解析命令並呈現輸出，`project`、`change`、`task`、`specdoc`、`archive` 各自處理一種本機資料責任。所有操作直接讀寫專案內的 Markdown/YAML，不呼叫 AI、Server 或 Git。

**技術：** Go 標準函式庫；以 `go test` 驗證套件與 CLI 行為。維持單一可執行檔，不引入資料庫或執行期服務。

**設計稿：** [2026-11-26-speclet-local-cli-design.md](../specs/2026-11-26-speclet-local-cli-design.md)

## 全域限制

- 專案資料根目錄名稱為 `speclet/`，內容為 UTF-8 Markdown 或 YAML。
- `speclet init [PATH]` 只初始化指定目錄（預設目前目錄）；其他命令由目前目錄向上尋找最近的 `speclet/`。
- 正式規格用 `## Requirement: <name>` 與 `### Scenario: <name>`；delta 用 `## ADDED|MODIFIED|REMOVED`、`### Requirement: <name>` 與 `#### Scenario: <name>`。
- 只解析標記需求、情境與平面 checkbox；一般 Markdown 和程式碼圍欄原文保留；巢狀任務拒絕處理。
- 成功訊息寫 stdout、錯誤訊息寫 stderr；成功結束碼為 0，錯誤使用非零結束碼。人類可讀輸出是唯一格式。
- 初始化、任務更新及封存遇到輸入錯誤時，不可覆寫原有內容。
- MVP 不管理 Git 操作、遠端服務、帳號、桌面 UI、AI 模型、worktree 或品質關卡。
- 首版以 macOS 可用為目標；路徑處理不得硬編碼單一作業系統分隔符。

## 檔案責任

| 路徑 | 責任 |
| --- | --- |
| `go.mod` | Go module 與最低語言版本 |
| `cmd/speclet/main.go` | 把作業系統參數、標準輸入輸出交給 CLI runner |
| `internal/cli/run.go` | 命令解析、專案目錄定位、輸出和結束碼 |
| `internal/project/layout.go` | 初始化佈局、驗證既有佈局、向上尋找專案根 |
| `internal/change/store.go` | 建立、列出、讀取進行中變更 |
| `internal/task/checkbox.go` | 解析任務 ID 與更新指定 checkbox |
| `internal/specdoc/parse.go` | 保留 Markdown 原文並辨識正式需求區塊 |
| `internal/specdoc/delta.go` | 驗證 delta、套用新增/修改/移除與標題正規化 |
| `internal/archive/archive.go` | 封存前完整驗證、更新正式規格、移動變更目錄 |

## Review Focus

- **路徑與名稱含空格或 Unicode**：操作必須留在專案目錄並保留名稱；Task 1、2 的測試覆蓋。
- **名稱含 `..`、絕對路徑或路徑分隔符**：拒絕操作且不得寫出 `speclet/` 外；Task 2 的測試覆蓋。
- **程式碼圍欄含看似 Requirement、Scenario 或 checkbox 的文字**：不得誤解析或改寫；Task 4 的測試覆蓋。
- **目前目錄內有巢狀 `speclet/`**：命令採最近的資料根；Task 1 的測試覆蓋。
- **封存目標已存在或正式規格路徑型態錯誤**：拒絕封存並保留原規格與進行中變更；Task 5 的測試覆蓋。

---

### Task 1：CLI 入口、專案初始化與根目錄搜尋

**檔案：**
- 新增：`go.mod`
- 新增：`cmd/speclet/main.go`
- 新增：`internal/cli/run.go`
- 新增：`internal/project/layout.go`
- 新增：`internal/project/layout_test.go`
- 新增：`internal/cli/run_test.go`

**介面：**
- 提供：`func Run(args []string, stdout, stderr io.Writer, cwd string) int`
- 提供：`func FindRoot(start string) (string, error)`，成功時回傳 `<專案目錄>/speclet` 的絕對路徑。
- 提供：`func Init(target string) error`，`target` 是要初始化的專案目錄，函式在其中建立 `speclet/`。
- CLI entry 只負責傳入 `os.Args[1:]`、標準串流與目前目錄，並以 `Run` 的回傳值結束程序。有效佈局必須有 `config.yaml`（內容含 `version: 1`）、`specs/`、`changes/` 與 `changes/archive/`。向上搜尋遇到最近的 `speclet/` 但結構無效時直接報錯，不跳過它改用外層專案。

- [x] **步驟 1：先寫 `FindRoot` 與 `Init` 的失敗測試**：無 `speclet/` 時回傳可辨識錯誤；既有但無效的 `speclet/` 初始化拒絕，原檔案內容不變。
- [x] **步驟 2：執行 `go test ./internal/project` 確認測試失敗。**
- [x] **步驟 3：實作向上搜尋**：由 `start` 逐層檢查到檔案系統根目錄，第一個有效 `speclet/` 即為結果；巢狀資料夾測試最近者優先。
- [x] **步驟 4：先寫初始化成功與重跑測試**：建立 `config.yaml`、`specs/`、`changes/`、`changes/archive/`；重跑不得改寫既有檔案；涵蓋帶空格與 Unicode 的目標路徑。
- [x] **步驟 5：執行 `go test ./internal/project` 確認新測試失敗，再實作 `Init`。**
- [x] **步驟 6：先寫 CLI 的 init、無效命令、stdout/stderr 與結束碼測試，再實作 `Run` 的最小 dispatcher。**
- [x] **步驟 7：執行 `go test ./...` 確認 Task 1 通過。**
- [ ] **步驟 8：提交本任務。**（依使用者指示略過，不提交）

### Task 2：變更建立、列表與檢視

**檔案：**
- 新增：`internal/change/store.go`
- 新增：`internal/change/store_test.go`
- 新增：`internal/task/checkbox.go`（先提供 Count）
- 新增：`internal/task/checkbox_test.go`
- 修改：`internal/cli/run.go`
- 修改：`internal/cli/run_test.go`

**介面：**
- 提供：`func Create(root, name string) error`
- 提供：`func List(root string) ([]Summary, error)`，`Summary` 欄位為 `Name string`、`Done int`、`Total int`；使用 `task.Count` 統計任務。
- 提供：`func Count(src []byte) (done, total int, err error)`，只統計平面 checkbox，拒絕巢狀 checkbox。
- 提供：`func Read(root, name string) (Detail, error)`；`Detail` 欄位為 `Name string`、`ProposalPath string`、`TasksPath string`、`DeltaPaths []string`。
- 變更名稱保留使用者輸入，但必須是單一路徑元件；拒絕空值、`.`、`..`、絕對路徑、`/`、`\` 和控制字元；允許 Unicode 與空格。

- [x] **步驟 1：先寫建立、重複名稱、無變更空列表、列表任務統計與找不到變更測試。**
- [x] **步驟 2：執行 `go test ./internal/change` 確認測試失敗。**
- [x] **步驟 3：實作安全名稱驗證與 `Create`**：建立 `proposal.md`、`tasks.md`、`specs/`；不覆蓋既有目錄。
- [x] **步驟 4：先寫 `List`、`Read` 對空檔與格式錯誤的行為測試，再完成兩個讀取介面。**
- [x] **步驟 5：接上 `change new/list/show` CLI**：輸出變更名稱、完成數和相關檔案位置；錯誤走 stderr。
- [x] **步驟 6：加入 `../escape`、絕對路徑與路徑分隔符測試，確認 `speclet/` 外沒有新檔案。**
- [x] **步驟 7：執行 `go test ./...` 確認 Task 2 通過。**
- [ ] **步驟 8：提交本任務。**（依使用者指示略過，不提交）

### Task 3：任務 checkbox 更新

**檔案：**
- 新增：`internal/task/checkbox.go`
- 新增：`internal/task/checkbox_test.go`
- 修改：`internal/change/store.go`
- 修改：`internal/cli/run.go`
- 修改：`internal/cli/run_test.go`

**介面：**
- 提供：`func SetDone(tasksPath, id string, done bool) error`
- 任務 ID 行格式為 `- [ ] T001 <文字>` 或 `- [x] T001 <文字>`；只更新精確 ID 的 checkbox，保留行內其他文字與檔案其他位元組。

- [x] **步驟 1：先寫完成、取消、未知 ID、重複 ID、巢狀 checkbox、空白行和 CRLF 保留測試。**
- [x] **步驟 2：執行 `go test ./internal/task` 確認測試失敗。**
- [x] **步驟 3：實作平面任務掃描與單行更新**：未知、重複或巢狀 ID 都回錯且不改檔。
- [x] **步驟 4：接上 `task done/undone` CLI**：操作前確認 change 與 tasks.md 存在；輸出變更名和任務 ID。
- [x] **步驟 5：執行 `go test ./...` 確認 Task 3 通過。**
- [ ] **步驟 6：提交本任務。**（依使用者指示略過，不提交）

### Task 4：正式規格與 delta Markdown 解析

**檔案：**
- 新增：`internal/specdoc/parse.go`
- 新增：`internal/specdoc/delta.go`
- 新增：`internal/specdoc/parse_test.go`
- 新增：`internal/specdoc/delta_test.go`

**介面：**
- 提供：`type Requirement struct { Name string; Raw []byte }` 與 `type Document struct { Prefix []byte; Requirements []Requirement; Suffix []byte }`；Requirement.Raw 保留區塊原文。
- 提供：`func ParseSpec(src []byte) (Document, error)`，保留原始行片段與可定位的 requirement 區塊。
- 提供：`type OperationKind int`，常數 `Added`、`Modified`、`Removed`；`type Operation struct { Kind OperationKind; Name string; Raw []byte }`。
- 提供：`func ParseDelta(src []byte) ([]Operation, error)`，每項操作帶唯一名稱和原始內容行。
- 提供：`func ApplyDelta(spec []byte, delta []byte) ([]byte, error)`，只修改指定 requirement，保留未修改的原文。

- [x] **步驟 1：先寫需求名稱唯一、情境層級、程式碼圍欄忽略 marker、普通前言/表格保留、無效標題與重複需求測試。**
- [x] **步驟 2：執行 `go test ./internal/specdoc` 確認測試失敗。**
- [x] **步驟 3：實作正式規格的行式解析**：只將程式碼圍欄外 `## Requirement:` 和 `### Scenario:` 視為結構標記；其他普通 Markdown 保留。
- [x] **步驟 4：先寫 delta 區段、需求層級、`#### Scenario`、未知區段、重複/缺失名稱測試，再實作 `ParseDelta`。**
- [x] **步驟 5：先寫 Added、Modified、Removed 輸出及原文保留測試，再實作 `ApplyDelta`；輸出正式規格時將 delta `#### Scenario` 轉為 `### Scenario`。**
- [x] **步驟 6：執行 `go test ./internal/specdoc` 確認 Task 4 通過。**
- [ ] **步驟 7：提交本任務。**（依使用者指示略過，不提交）

### Task 5：封存與規格更新

**檔案：**
- 新增：`internal/archive/archive.go`
- 新增：`internal/archive/archive_test.go`
- 修改：`internal/cli/run.go`
- 修改：`internal/cli/run_test.go`

**介面：**
- 提供：`func Run(specletRoot, name string) error`，`specletRoot` 是 `<專案目錄>/speclet`。
- `Run` 先載入並驗證整個變更及所有目標規格，再寫入正式規格，最後將變更目錄移到 `changes/archive/YYYY-MM-DD-name/`。

- [x] **步驟 1：先寫全部任務已完成的 Added/Modified/Removed 封存成功測試，確認正式規格內容與 archive 目錄內容。**
- [x] **步驟 2：執行 `go test ./internal/archive` 確認測試失敗。**
- [x] **步驟 3：實作任務完成度檢查**：未完成任務拒絕封存；零任務視為完成。
- [x] **步驟 4：先寫 delta 不合法、目標需求不存在或重複、重複封存路徑、正式規格路徑是目錄等失敗測試，斷言原規格與 change 檔案未改。**
- [x] **步驟 5：實作全量預檢與封存**：全部 delta 驗證成功後才準備寫入；同名封存目錄不覆蓋；寫入採同檔案系統暫存與 rename，失敗時復原已替換的規格檔並保留 change。
- [x] **步驟 6：接上 `archive` CLI**：成功 stdout 摘要；失敗 stderr 說明原因並回傳非零。
- [x] **步驟 7：執行 `go test ./...` 確認 Task 5 通過。**
- [ ] **步驟 8：提交本任務。**（依使用者指示略過，不提交）

### Task 6：端到端驗收與使用說明

**檔案：**
- 修改：`README.md`
- 新增：`internal/cli/workflow_test.go`

**介面：**
- `workflow_test.go` 以 `Run` 和暫存專案目錄覆蓋完整 CLI 旅程，不另啟動外部服務。

- [x] **步驟 1：先寫端到端測試**：init → change new → 寫入 proposal/tasks/delta → list/show → task done → archive，斷言最終正式規格、封存資料與 stdout/stderr/結束碼。
- [x] **步驟 2：執行 `go test ./...` 確認端到端測試失敗或暴露缺口。**
- [x] **步驟 3：補齊 CLI help 與 README 使用範例**：以繁體中文說明 Go 安裝/建置、init、建立變更、任務更新、delta 格式與 archive 守門。
- [x] **步驟 4：執行 `go test ./...` 確認全套測試通過。**
- [x] **步驟 5：執行 `go build -o /tmp/speclet ./cmd/speclet`，並以建置出的 CLI 在暫存目錄完成 init 與錯誤命令驗收。**
- [ ] **步驟 6：提交本任務。**（依使用者指示略過，不提交）
