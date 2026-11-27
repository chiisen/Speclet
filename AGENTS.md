# Speclet 專案 Agent 指南

## 功能開發時使用 Speclet

實作新的使用者可見功能或變更既有行為時，使用本 repo 的 Speclet CLI 追蹤工作：

1. 閱讀 repo README 的安裝與 PATH 設定說明，並用 `command -v speclet` 確認 CLI 可用。
2. 如果 repo 尚無 `speclet/` 資料目錄，從 repo 根目錄執行 `speclet init`。若初始化失敗或既有目錄結構無效，停止操作並回報錯誤，不要覆寫專案資料。
3. 使用 `speclet change new <name>` 建立變更，在 `proposal.md` 記錄簡要理由、在 `tasks.md` 列出可驗收待辦，並在每個受影響功能的 `specs/<capability>/spec.md` 記錄規格差異。
4. 實作變更，並用 `speclet task done` 或 `speclet task undone` 維護待辦狀態。行為變更需新增或更新測試，並執行相關檢查。
5. 完成前使用 `speclet change show <name>` 和 `speclet change list` 檢視變更。只有所有待辦完成，且規格差異描述已交付行為時，才執行 `speclet archive <name>`。
6. 最後回報 Speclet change 名稱、封存結果，以及未能執行的檢查。Speclet 管理本機規格檔案；Git commit、branch 和 push 依使用者指示另行處理。

不改變產品行為的文件修訂與機械式維護，不需要建立 Speclet change。

## 處理 Speclet 任務中發現的缺陷

依照目前任務已確認的驗收條件與範圍，決定如何處理：

- **缺陷在目前範圍內，且影響驗收。** 在目前 change 中修復，重跑原本會失敗的案例，並回報結果；不需要另開 issue。
- **缺陷阻擋驗收，且修復需要擴大範圍。** 告知使用者阻擋原因與所需的額外修改，讓使用者決定是否納入本次工作；繼續進行不受影響的部分。
- **已確認的缺陷在目前任務範圍外，且不影響本次驗收。** 先檢查 `chiisen/Speclet` 是否已有涵蓋同一缺陷的 issue。有的話回報 issue 連結；沒有的話建立 GitHub issue，記錄受影響版本、最小重現步驟、預期與實際結果及影響。移除回報中的密鑰和個人資料。建立 issue 後繼續原任務，修復時程由使用者安排。若無法查重或建立 issue，提供可直接提交的草稿並說明原因。
- **尚未確認的現象或改進想法。** 若有具體影響，在任務回報中說明；只有已確認的缺陷才建立 bug issue。

日後修復已追蹤的缺陷時，記錄重跑原案例的結果。若修復已驗證，且 PR 的目標分支是預設分支，使用 `Fixes #<issue-number>`，讓 issue 在 PR 合併時關閉。
