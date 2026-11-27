## ADDED

### Requirement: 提供腳本可逐行讀取的變更名稱

`speclet change list --names-only` 沿用 `speclet change list` 的變更範圍、排序和驗證規則，輸出只有變更名稱。

#### Scenario: 有進行中的變更

- **WHEN** 使用者以 `--names-only` 列出進行中的變更
- **THEN** stdout 依原本的排序，每行輸出一個完整名稱，並保留名稱中的空格

#### Scenario: 沒有進行中的變更

- **WHEN** 使用者在沒有進行中變更的已初始化專案執行 `speclet change list --names-only`
- **THEN** 命令成功結束，stdout 為空

#### Scenario: 沿用預設列表格式

- **WHEN** 使用者直接執行 `speclet change list`
- **THEN** 列表仍顯示各變更的任務進度，空清單時仍顯示 `No active changes.`
