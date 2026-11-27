## Requirement: 設定 CLI 顯示語系

### Scenario: 預設語系

- **WHEN** 電腦尚未設定 Speclet 顯示語系
- **THEN** CLI 使用英文

### Scenario: 列出和查看語系

- **WHEN** 使用者執行語系列表命令
- **THEN** CLI 列出所有已提供翻譯的語系，包含 `en` 和 `zh-TW`

- **WHEN** 使用者查看目前語系
- **THEN** CLI 顯示目前使用的語系；尚未設定時顯示英文為預設值

### Scenario: 保存偏好

- **WHEN** 使用者將語系設為支援的選項
- **THEN** Speclet 將偏好保存在該電腦的使用者設定目錄，之後執行命令時沿用

### Scenario: 拒絕不支援的語系

- **WHEN** 使用者設定未列出的語系
- **THEN** CLI 不變更目前偏好，並顯示可選語系

## Requirement: 依偏好顯示 CLI 文字

### Scenario: 使用繁體中文

- **WHEN** 目前語系為 `zh-TW`
- **THEN** CLI 的說明、狀態、錯誤和診斷文字使用繁體中文

### Scenario: 保持資料與機器輸出格式

- **WHEN** 使用者以 `--names-only` 查詢變更名稱，或讀寫規格標記
- **THEN** 輸出名稱、命令、檔案路徑和規格標記不因顯示語系而改變
