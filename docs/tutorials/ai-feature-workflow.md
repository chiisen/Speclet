# 使用 Speclet 開發功能

這份教學示範 AI coding agent 如何依照 `AGENTS.md`，從建立規格變更、測試先行、實作、更新待辦，一路完成驗收和封存。CLI 的第一次操作請先看[建立第一個 Speclet 變更](first-change.md)。

以下用 `speclet change list --names-only` 作為範例。它會逐行輸出進行中的變更名稱，方便 shell 腳本讀取。

## 1. 確認專案和工作規範

先在 repository 根目錄查看 `AGENTS.md`、README 和 Git 狀態，確認目前工作的範圍，並保留原有修改。確認 CLI 可執行：

```sh
command -v speclet
speclet help
```

如果這個 repository 還沒有 `speclet/` 資料目錄，在根目錄執行 `speclet init`。若已有資料目錄，就先檢查它是否有效，不要初始化覆寫。

當程式碼剛修改但 PATH 上的 `speclet` 還是舊版時，使用目前原始碼執行 `go run -buildvcs=true ./cmd/speclet ...`，或重新建置並執行新 binary。不要用舊版 CLI 的結果判斷新功能。

## 2. 建立 Speclet 變更

變更名稱使用簡短、能辨認工作的英文 kebab-case。提案、待辦、需求名稱和情境內文都可以使用繁體中文。

```sh
speclet change new change-list-names-only
speclet change show change-list-names-only
```

編輯 `proposal.md`，記錄原因和預期行為；在 `tasks.md` 寫下可驗收的工作，為每項工作指定唯一 ID；在 `specs/<capability>/spec.md` 寫入需求差異。Speclet 會辨識固定標記，因此保留 `ADDED`、`MODIFIED`、`REMOVED`、`Requirement:`、`Scenario:`、`WHEN`、`THEN` 和平面 checkbox 格式。

例如，清單功能的待辦可寫成：

```markdown
## 待辦事項

- [ ] T001 新增只列名稱、空清單和無效參數的 CLI 測試
- [ ] T002 實作 `--names-only` 並確認原有輸出不變
- [ ] T003 更新 CLI help 和 README
- [ ] T004 執行完整測試並驗收建置出的 CLI
```

建立變更後，用 `speclet change list` 確認任務總數正確。

## 3. 先測試，再實作

每次只完成一個可驗收的待辦。先新增能描述預期行為的測試，執行測試並確認失敗原因是功能尚未實作，再寫足以通過測試的程式碼。

```sh
go test ./internal/cli -run '^TestRunChangeListNamesOnly$' -count=1 -v
```

測試有預期的失敗結果後，再實作功能並重跑同一測試。通過後，以 `speclet task done <change> <task-id>` 更新該待辦。

```sh
speclet task done change-list-names-only T001
speclet change list
```

若驗收仍未完成，保留未勾選狀態；若曾提早勾選，可用 `speclet task undone <change> <task-id>` 改回未完成。每次更新後檢查 `speclet change list` 顯示的完成數。

## 4. 處理開發中發現的缺陷

用目前任務的範圍和驗收條件判斷：

- **缺陷與目前功能相關。** 在這次變更內修復，重跑原本能重現問題的案例，再繼續任務。
- **缺陷已確認、與目前功能無關，而且不影響本次驗收。** 先查 `chiisen/Speclet` 是否已有涵蓋同一問題的 Issue。有就回報原 Issue，沒有就建立 Issue，記錄受影響版本、最小重現步驟、預期結果、實際結果和影響。建立後繼續原任務，修復時間由使用者安排。
- **缺陷阻擋驗收，且修復需要擴大範圍。** 回報阻擋原因和所需變更，讓使用者決定是否納入本次。可以先完成其他不受影響的工作。
- **現象尚未確認。** 記錄觀察和不確定處，不把推測當成已確認缺陷。回報中不得包含密鑰或個人資料。

## 5. 完成驗收、檢視並封存

執行完整測試和 CLI 驗收。清單功能可以檢查有空格的名稱、排序、空清單、預設輸出和錯誤參數：

```sh
go test ./...
go build -o /tmp/speclet ./cmd/speclet
```

接著明確使用剛建置的 `/tmp/speclet`，避免 PATH 上的舊版遮蔽本次結果。在暫存專案中初始化並建立一個有空格名稱的變更：

```sh
demo_dir=$(mktemp -d "${TMPDIR:-/tmp}/speclet-check.XXXXXX")
(
  cd "$demo_dir" &&
    /tmp/speclet init &&
    /tmp/speclet change new "sample change" &&
    /tmp/speclet change list --names-only
)
```

清單應只輸出一行 `sample change`。上面的子 shell 會保留目前終端機所在位置；完成實際變更的待辦後，在原專案根目錄檢視並封存本次 change：

```sh
/tmp/speclet change show change-list-names-only
/tmp/speclet change list
/tmp/speclet archive change-list-names-only
/tmp/speclet change list
```

封存後，`change list` 應顯示 `No active changes.`，正式規格應包含已封存的需求。

## 6. 回顧 `AGENTS.md` 是否有效

完成時簡短回顧哪些規則有實際執行、遇到哪些卡點，以及哪些情境沒有被測到。例如，若這次沒有發現範圍外缺陷，就明確記錄 Issue 分流規則尚未實測。只提出有實際依據的改進建議。

一般 coding 任務由 agent 在完成驗證後自行更新 Speclet 待辦。若這是一場使用者親自練習 CLI 的教學，agent 可在每個階段提供下一條命令，讓使用者親自執行並確認結果。
