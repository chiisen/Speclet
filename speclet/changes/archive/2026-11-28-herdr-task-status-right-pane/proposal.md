# Change: herdr-task-status-right-pane

## Why

目前 Herdr 整合以 popup 顯示 Speclet 狀態，會遮住終端工作區，也必須反覆開啟才能查看。將狀態放在目前 tab 右側，能讓進度持續可見。

## What Changes

- 從 Herdr 指令面板以右側分割 pane 開啟 `speclet-status`，取代 popup。
- 保留唯讀狀態、專案切換、更新頻率、手動重新整理及錯誤提示。
- 更新整合文件與驗收規格。

追蹤：chiisen/Speclet#11。
