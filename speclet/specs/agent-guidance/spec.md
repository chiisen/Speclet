## Requirement: AI coding agent repository instructions

### Scenario: 使用繁體中文閱讀專案規則

- **WHEN** AI coding agent 閱讀 repo 根目錄的 `AGENTS.md`
- **THEN** 規則說明使用繁體中文，命令、路徑、程式識別字和固定英文標記維持可直接使用的原文

### Scenario: 維持既有規範行為

- **WHEN** AI coding agent 依照 `AGENTS.md` 執行專案任務
- **THEN** 各項規則的意義、優先順序和適用範圍與改寫前一致
