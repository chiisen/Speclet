## Requirement: 進行中變更的機器可讀摘要

CLI 必須提供 `speclet change list --json`，以穩定的機器可讀格式列出進行中的變更。每個陣列項目必須使用小駝峰欄位 `name`、`done` 和 `total`；任務數必須沿用既有的任務計數結果。沒有進行中變更時必須輸出 `[]`。此模式的 stdout 只能包含 JSON；參數無效或互相衝突時，必須將錯誤寫入 stderr 並以非零狀態結束。預設的人類可讀輸出及 `--names-only` 行為必須維持不變。

### Scenario: 以 JSON 列出進行中的變更

- **WHEN** 在已初始化且有進行中變更的專案執行 `speclet change list --json`
- **THEN** stdout 輸出包含正確任務數的進行中變更摘要 JSON 陣列

### Scenario: 以 JSON 列出空專案

- **WHEN** 專案沒有進行中的變更時執行 `speclet change list --json`
- **THEN** stdout 輸出 `[]`，並以換行結尾

### Scenario: 拒絕衝突或不支援的列表參數

- **WHEN** 同時使用 `--json` 與 `--names-only`，或提供不支援的列表參數
- **THEN** CLI 將用法錯誤寫入 stderr、不輸出 stdout，並以非零狀態結束

## Requirement: 選用的唯讀 Herdr 任務狀態面板

本 repo 必須在 `integrations/herdr/` 提供選用的 Herdr mod。Herdr 指令面板必須在目前 tab 的焦點 pane 右側開啟此 mod，使用一般分割 pane，不使用 popup 或 overlay；pane 可由 Herdr 調整、關閉及再次開啟。此 mod 必須判定目前 Herdr 專案的目錄與機器，使用參數陣列呼叫 `speclet` 可執行檔，並以該目錄作為工作目錄；面板必須顯示所有進行中變更的名稱及已完成／總任務數。此 mod 只能讀取專案狀態。未初始化的專案、找不到 CLI、CLI 執行失敗、JSON 無效，以及已初始化但沒有進行中變更，必須各自有明確提示。切換專案時，必須清除前一個專案的狀態並載入新專案的狀態。

### Scenario: 顯示進行中的任務進度

- **WHEN** 目前專案有進行中的 Speclet 變更，且 Speclet CLI 執行成功
- **THEN** 右側 pane 顯示每個變更的名稱及已完成／總任務數

### Scenario: 顯示空清單、未初始化及錯誤狀態

- **WHEN** 專案沒有進行中變更、尚未初始化、找不到 CLI、CLI 執行失敗或輸出無效 JSON
- **THEN** pane 以清楚且不同的方式呈現狀態，不得將 CLI 錯誤顯示成空清單

### Scenario: 切換專案

- **WHEN** Herdr 切換目前專案
- **THEN** 右側 pane 捨棄前一個專案的資料，並重新載入新專案的狀態
