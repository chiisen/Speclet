# 建立第一個 Speclet 變更

這次用一個會印出 `你好，Speclet！` 的小腳本練習。你會先寫下要做什麼，再完成兩項待辦，最後把驗證過的行為更新到正式規格。

開始前先確認 `speclet help` 可以執行。若終端機找不到指令，請參考 [README 安裝說明](../../README.md#安裝)。

這份教學介紹 Speclet CLI 的基本操作。若要了解 AI coding agent 如何把 Speclet 用在功能開發，請看[使用 Speclet 開發功能](ai-feature-workflow.md)。

## 1. 初始化練習專案

開一個獨立的暫存目錄，在裡面執行 `speclet init`。

```sh
demo_dir=$(mktemp -d "${TMPDIR:-/tmp}/speclet-demo.XXXXXX") &&
  cd "$demo_dir" &&
  speclet init
```

看到 `Initialized Speclet project.` 後，目錄裡會有 `speclet/config.yaml`、存放正式規格的 `speclet/specs/`，以及存放變更與封存資料的 `speclet/changes/`。同一個專案只需初始化一次。

## 2. 建立變更並找到檔案

```sh
speclet change new add-greeting
speclet change show add-greeting
```

`change new` 會建立以下骨架，`change show` 會列出檔案路徑。

```text
speclet/changes/add-greeting/
├── proposal.md  有基本標題的提案
├── tasks.md     空的待辦檔
└── specs/      空的規格差異目錄
```

## 3. 填寫提案、待辦和規格差異

用編輯器打開 `speclet/changes/add-greeting/proposal.md`，寫下變更原因和範圍。可以直接使用下面的內容。

```markdown
# 新增歡迎訊息

## 為什麼要改

第一次執行範例時，使用者需要看到一則訊息，確認腳本已正常執行。

## 變更內容

新增 `greet.sh`，執行時輸出一行 `你好，Speclet！`。
```

在同一個變更的 `tasks.md` 加入兩項待辦。每一項用平面 checkbox 和唯一的 ID，之後由 CLI 按 ID 更新勾選狀態。

```markdown
## 待辦事項

- [ ] T001 建立歡迎訊息腳本
- [ ] T002 執行腳本並確認輸出
```

建立規格差異的功能目錄。

```sh
mkdir -p speclet/changes/add-greeting/specs/greeting
```

在 `speclet/changes/add-greeting/specs/greeting/spec.md` 寫入以下內容。這裡是新增需求，所以使用 `## ADDED`。標記後的需求名稱、情境名稱和內文可以寫繁體中文。

```markdown
## ADDED

### Requirement: 顯示歡迎訊息

#### Scenario: 執行歡迎腳本

- **WHEN** 使用者執行 `sh greet.sh`
- **THEN** 終端機輸出一行 `你好，Speclet！`
```

執行 `speclet change list`，應看到 `add-greeting (0/2 tasks done)`。

## 4. 完成工作並更新待辦

在練習專案根目錄建立 `greet.sh`，然後標記第一項完成。

```sh
cat > greet.sh <<'EOF'
#!/bin/sh
printf '%s\n' '你好，Speclet！'
EOF
speclet task done add-greeting T001
speclet change list
```

此時應顯示 `1/2 tasks done`。接著執行腳本。

```sh
sh greet.sh
```

看到一行 `你好，Speclet！` 後，再勾選第二項。

```sh
speclet task done add-greeting T002
speclet change list
```

確認清單顯示 `2/2 tasks done`。如果提早勾選了，可以用 `speclet task undone add-greeting T002` 恢復待辦狀態。

## 5. 封存並檢查結果

```sh
speclet archive add-greeting
speclet change list
cat speclet/specs/greeting/spec.md
ls speclet/changes/archive
```

封存後會看到三個結果。

- `change list` 顯示 `No active changes.`。
- `speclet/specs/greeting/spec.md` 出現正式需求 `## Requirement: 顯示歡迎訊息`，以及 `### Scenario: 執行歡迎腳本`。
- `speclet/changes/archive/` 下出現 `YYYY-MM-DD-add-greeting`，保留這次的提案、待辦和規格差異。

若有待辦未完成或規格差異格式有誤，`archive` 會指出原因，變更會留在進行中清單。修正後可重新執行封存。
