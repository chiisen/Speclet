## ADDED

### Requirement: CLI 應用程式版本資訊

Speclet 執行檔必須提供 `speclet --version`，並在 `speclet -h`／`speclet --help` 顯示應用程式版本。正式建置使用 linker 注入的語意化 Git tag，例如 `v0.1.0`；未注入 tag 的開發建置顯示 `dev+` 與 Go build info 中 VCS revision 的前七個字元。若 build info 沒有 revision，必須顯示 `dev`。版本查詢不得執行 Git 或依賴 Speclet 專案資料；Go toolchain 版本及 `speclet/config.yaml` 格式版本不得冒充應用程式版本。

#### Scenario: 查詢正式版本

- **WHEN** 執行檔以 linker 注入 `v0.1.0` 建置並執行 `speclet --version`
- **THEN** 輸出包含 `v0.1.0`

#### Scenario: 查詢開發版本

- **WHEN** 執行一般 Go 開發建置並執行 `speclet --version`
- **THEN** 輸出包含 `dev+` 及 VCS revision 前七個字元

#### Scenario: Help 顯示版本

- **WHEN** 執行 `speclet -h` 或 `speclet --help`
- **THEN** help 輸出包含目前執行檔版本及既有命令說明

#### Scenario: Build info 沒有 VCS revision

- **WHEN** 開發建置未帶 tag 且 build info 不含 VCS revision
- **THEN** `speclet --version` 顯示 `dev`，不發生錯誤
