# Speclet 專案 Agent 指南

## 功能開發使用 Speclet

新增使用者可見功能或修改既有行為時，請用本 repo 的 Speclet CLI 記錄工作進度：

1. 閱讀 README 的安裝與 PATH 設定說明，並用 `command -v speclet` 確認 CLI 可用。
2. 如果專案還沒有 `speclet/` 資料目錄，請在專案根目錄執行 `speclet init`。若初始化失敗或目錄結構無效，先停止並回報錯誤，不要覆寫現有資料。
3. 使用 `speclet change new <name>` 建立變更。在 `proposal.md` 說明原因、於 `tasks.md` 列出可驗收的待辦，並在每個受影響功能的 `specs/<capability>/spec.md` 記錄規格差異。
4. 實作時，用 `speclet task done` 或 `speclet task undone` 更新待辦狀態。若改變程式行為，請新增或更新測試，並執行相關檢查。
5. 工作完成前，用 `speclet change show <name>` 和 `speclet change list` 檢查變更。所有待辦完成、規格差異也符合實際交付內容後，才執行 `speclet archive <name>`。
6. 完成時回報 Speclet change 名稱、是否已封存，以及哪些檢查未能執行。Speclet 管理本機規格檔案；Git commit、branch 和 push 依使用者指示另行處理。

文件修改或例行維護若不改變產品行為，就不必建立 Speclet change。

## 處理 Speclet 工作中發現的缺陷

按照目前任務已確認的範圍與驗收條件處理：

- **缺陷在目前範圍內，且影響驗收。** 在這次 change 中修復，重跑原本會失敗的案例，並回報結果；不必另開 issue。
- **缺陷擋住驗收，修復又需要擴大範圍。** 說明目前卡住的部分和所需修改，讓使用者決定是否納入本次工作。其他不受影響的工作可以繼續。
- **已確認的缺陷與本次無關，也不影響驗收。** 先查 `chiisen/Speclet` 是否已有相同缺陷的 issue。有的話回報連結；沒有的話建立 GitHub issue，寫明受影響版本、最小重現步驟、預期與實際結果，以及影響。不要放入密鑰或個人資料。建立 issue 後繼續原任務，修復時程由使用者決定。如果無法查重或建立 issue，附上可直接提交的草稿，並說明原因。
- **還沒確認的現象或改進想法。** 若有具體影響，寫進任務回報。確認是缺陷後才開 bug issue。

之後修復已追蹤的缺陷時，記錄重跑原案例的結果。修復已驗證且 PR 以預設分支為目標時，在 PR 使用 `Fixes #<issue-number>`，讓 issue 在合併時關閉。
