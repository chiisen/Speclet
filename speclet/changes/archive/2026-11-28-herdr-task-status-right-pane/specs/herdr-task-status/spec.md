## MODIFIED

### Requirement: 選用的唯讀 Herdr 任務狀態面板

本 repo 必須在 `integrations/herdr/` 提供選用的 Herdr mod。Herdr 指令面板必須在目前 tab 的焦點 pane 右側開啟此 mod，使用一般分割 pane，不使用 popup 或 overlay；pane 可由 Herdr 調整、關閉及再次開啟。此 mod 必須判定目前 Herdr 專案的目錄與機器，使用參數陣列呼叫 `speclet` 可執行檔，並以該目錄作為工作目錄；面板必須顯示所有進行中變更的名稱及已完成／總任務數。此 mod 只能讀取專案狀態。未初始化的專案、找不到 CLI、CLI 執行失敗、JSON 無效，以及已初始化但沒有進行中變更，必須各自有明確提示。切換專案時，必須清除前一個專案的狀態並載入新專案的狀態。

#### Scenario: 顯示進行中的任務進度

- **WHEN** 目前專案有進行中的 Speclet 變更，且 Speclet CLI 執行成功
- **THEN** 右側 pane 顯示每個變更的名稱及已完成／總任務數

#### Scenario: 顯示空清單、未初始化及錯誤狀態

- **WHEN** 專案沒有進行中變更、尚未初始化、找不到 CLI、CLI 執行失敗或輸出無效 JSON
- **THEN** pane 以清楚且不同的方式呈現狀態，不得將 CLI 錯誤顯示成空清單

#### Scenario: 切換專案

- **WHEN** Herdr 切換目前專案
- **THEN** 右側 pane 捨棄前一個專案的資料，並重新載入新專案的狀態
