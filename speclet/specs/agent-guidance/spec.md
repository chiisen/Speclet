## Requirement: AI coding agent repository instructions

### Scenario: 繁體中文指南保留原有指令

- **WHEN** AI coding agent 閱讀 repo 根目錄的 `AGENTS.md`
- **THEN** 說明使用繁體中文，命令、路徑、程式識別字和固定標記維持原文

### Scenario: 翻譯不改變規則

- **WHEN** AI coding agent 依照 `AGENTS.md` 執行任務
- **THEN** 規則的意義、優先順序和適用範圍與改寫前相同
