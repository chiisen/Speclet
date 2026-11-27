# Herdr 任務狀態右側面板設計與實作計畫

> **執行代理必要子技能：** 使用 `superpowers:executing-plans` 逐項執行本計畫。

**目標：** 在目前 Herdr 分頁右側開啟常駐的 Speclet 任務狀態面板。

**架構：** 將指令面板命令改為 Herdr 的 `pane.split`，參數設為 `direction: "right"` 與 `mod: "speclet-status"`。移除 popup 專用的 manifest 設定和面板中的浮層關閉按鈕，保留現有狀態模型及重新整理行為。

**技術：** Herdr mod JSON、JavaScript、Node.js 內建測試執行器。

**設計依據：** 本文件的「設計決策」章節；[GitHub issue #11](https://github.com/chiisen/Speclet/issues/11)。

## 設計決策

目前任務狀態以 popup 顯示，會暫時遮住終端工作區；需求是讓狀態留在右側，工作時持續可見。

採用目前 Herdr 分頁的右側分割 pane。終端仍作為主要工作區，使用者可調整 pane 寬度，並透過 Herdr 的 pane 控制關閉或再次開啟。另一方案是整個 space 共用的側欄，切換分頁時仍可使用；它需要整合 Herdr layout 層，超出本次範圍。

狀態資料維持唯讀，沿用目前專案追蹤、定時與手動更新，以及錯誤狀態提示；不改變 CLI 或任務狀態資料格式。

後續章節只記錄實作步驟、檔案、測試及驗收方式。

## 全域限制

- 面板只能唯讀執行 `speclet change list --json`。
- 保留目前的專案追蹤、每 10 秒更新、手動重新整理，以及空清單、未初始化、找不到 CLI、CLI 錯誤和無效 JSON 的區別提示。
- 範圍限於目前分頁右側面板，不新增整個 space 共用的側欄。
- 沿用現有 Node 測試環境，不新增相依套件。

## 審查重點

- 指令面板必須把 mod 放在焦點面板右側；以 manifest 測試鎖定命令和參數。
- 不得保留 popup 設定；測試確認 manifest 不再有 `overlay` 欄位。
- 一般 pane 不應顯示浮層專用的關閉控制；移除條件式 UI，並在 Herdr 中確認。
- 空清單和錯誤結果仍須顯示不同訊息；保留現有狀態模型測試。
- 面板必須能在 Windows Herdr 中使用；執行 mod 檢查並確認實際面板。

---

### 任務 1：建立 Speclet 變更並記錄規格

**檔案：**
- 新增：`speclet/changes/herdr-task-status-right-pane/proposal.md`
- 新增：`speclet/changes/herdr-task-status-right-pane/tasks.md`
- 新增：`speclet/changes/herdr-task-status-right-pane/specs/herdr-task-status/spec.md`

**介面：**
- 使用：本文件的「設計決策」章節。
- 產出：明確記錄右側面板行為與可驗收工作的 Speclet change。

- [ ] 執行 `C:\Users\chiis\go\bin\speclet.exe change new herdr-task-status-right-pane` 建立變更。
- [ ] 更新提案、待辦與規格差異，限定於目前分頁右側面板，並列出測試、實作、文件和驗證工作。
- [ ] 執行 `speclet change show herdr-task-status-right-pane` 與 `speclet change list` 檢查結果。

### 任務 2：先以失敗測試鎖定 Herdr 面板位置

**檔案：**
- 修改：`integrations/herdr/status.test.js`
- 讀取：`integrations/herdr/mods/speclet-status/mod.json`

**介面：**
- 使用：現有 Node 測試執行器和 mod manifest。
- 產出：測試 `palette opens the status mod in a right split pane`，確認「Speclet：任務狀態」使用 `pane.split`，參數為 `{ direction: "right", mod: "speclet-status" }`，且 manifest 不含 `overlay` 欄位。

- [ ] 使用 `node:fs/promises` 讀取 `mod.json`，新增 manifest 行為測試。
- [ ] 在 `integrations/herdr` 執行 `node --test`，確認新測試因 manifest 仍使用 popup 而失敗。

### 任務 3：將 mod 改為右側 pane

**檔案：**
- 修改：`integrations/herdr/mods/speclet-status/mod.json`
- 修改：`integrations/herdr/mods/speclet-status/main.js`
- 修改：`integrations/herdr/README.md`
- 修改：`CHANGELOG.md`

**介面：**
- 使用：任務 2 的 manifest 測試。
- 產出：使用右側 `pane.split` 開啟 `speclet-status` 的指令；不含 popup 專用關閉行為的 pane；使用 `herdr-app do pane.split` 的跨平台說明。

- [ ] 修改 palette 命令及參數，移除 popup `overlay` 設定。
- [ ] 從 `main.js` 移除 `closeOverlay`、`props` 匯入及條件式關閉按鈕；保留重新整理和狀態呈現。
- [ ] 更新根目錄 README 與整合 README 的安裝／開啟說明，描述目前分頁右側面板行為，並提供 Windows PowerShell 與 macOS／Linux 範例。
- [ ] 在根目錄 `CHANGELOG.md` 加入簡短的繁體中文紀錄。
- [ ] 在 `integrations/herdr` 執行 `node --test`，確認全部測試通過。

### 任務 4：驗證並完成 Speclet change

**檔案：**
- 修改：`speclet/changes/herdr-task-status-right-pane/tasks.md`

**介面：**
- 使用：任務 1 至 3 的產出。
- 產出：已驗證並完成待辦的 Speclet change，以及更新後的 Herdr 安裝版 mod。

- [ ] 執行 `herdr-app check C:\Users\chiis\.config\herdr-app\mods\speclet-status`。
- [ ] 將更新後的 mod 複製到 Herdr 安裝目錄，並以右側方向執行 `herdr-app do pane.split` 開啟。
- [ ] 實際確認終端 pane 留在左側、任務狀態面板在右側；依目前環境檢查重新整理及空狀態／錯誤狀態。
- [ ] 執行 `git diff --check`，檢查 `speclet change show` 與 `speclet change list`，再把已完成工作標記完成。
- [ ] 僅在所有 Speclet 待辦完成且規格差異符合交付內容後封存 change。
