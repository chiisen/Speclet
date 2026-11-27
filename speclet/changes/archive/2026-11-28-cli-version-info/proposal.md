# Change: cli-version-info

## Why

[Issue #10](https://github.com/chiisen/Speclet/issues/10) 指出 CLI 目前無法顯示應用程式版本，導致使用者難以確認安裝版本或回報問題。Go toolchain 版本及 `speclet/config.yaml` 格式版本都不代表 Speclet 執行檔版本。

## What Changes

- 新增 `speclet --version`，並在 help 輸出版本資訊。
- 正式建置從語意化 Git tag 注入版本；未標記的開發建置顯示 `dev+` 與短 commit SHA。
- 以 Go build info 取得開發 revision，執行檔不需要在執行時呼叫 Git。
