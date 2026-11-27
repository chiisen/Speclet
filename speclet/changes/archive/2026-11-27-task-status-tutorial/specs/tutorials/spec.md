## ADDED

### Requirement: 說明 Speclet 任務狀態

#### Scenario: 查看任務完成進度

- **WHEN** 讀者依照簡易教學建立一項待辦並執行 `speclet change list`
- **THEN** 教學說明如何用 `speclet task done` 標記完成，並以清單顯示進度變化

#### Scenario: 取消完成標記

- **WHEN** 讀者把尚未完成的工作標成完成
- **THEN** 教學說明如何用 `speclet task undone` 還原狀態
