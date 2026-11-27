# 變更：herdr-task-status

## 原因

Speclet 以本機 Markdown 檔案追蹤功能工作，但使用者目前必須離開 Herdr、改用 CLI 查詢任務進度。新增精簡的選用 Herdr mod，讓使用者直接查看唯讀狀態，同時維持 Speclet 核心不依賴 Herdr。

## 變更內容

- 為 `speclet change list` 新增 `--json` 輸出，提供穩定的小駝峰欄位，並維持既有文字及 `--names-only` 格式。
- 新增選用的 Herdr 彈出面板，呼叫目前專案的 Speclet CLI，顯示進行中的變更與任務數。
- 補充安裝、開啟面板、CLI 需求與唯讀範圍的文件。
