# Speclet

Speclet 是一個本機優先的規格驅動開發 CLI，目標是在單一專案中，以容易閱讀的檔案管理規格、變更提案與任務清單。

## 專案狀態

MVP 已採 Go 實作，提供初始化、變更管理、任務追蹤和安全封存功能。

## 安裝

需要 Go 1.27 或更新版本。在 repo 根目錄執行：

```sh
go install ./cmd/speclet
```

確認 Go 的 `bin` 目錄已加入 `PATH`，即可執行 `speclet help`。也可以建置獨立執行檔：

```sh
go build -o /tmp/speclet ./cmd/speclet
```

## 快速開始

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

除 `init` 外，命令會從目前目錄向上尋找最近的 `speclet/`。`init` 只初始化指定路徑（預設目前目錄），不向上搜尋。錯誤寫入 stderr 並以非零結束碼退出；成功摘要寫入 stdout。

Speclet 不會操作 Git commit、branch 或 push，也不會呼叫 AI 模型。AI coding agent 可透過這些檔案和命令參與工作流程。
