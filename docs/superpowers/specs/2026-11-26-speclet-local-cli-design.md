# Speclet 本機 CLI MVP 設計

狀態：待使用者審核。本文是產品與行為設計，尚未開始實作。

## 意圖

Speclet 是給個人開發者及其 AI coding agent 使用的本機規格驅動開發工具。它讓專案以可讀的文字檔保存正式規格、待做變更與任務，並提供一組 CLI 動作管理這些資料。

成功標準是：開發者可以從終端機初始化資料夾、建立變更、查看待辦、完成任務，最後把變更中的規格更新併回正式規格並封存；整個流程不需要 Server 或桌面程式。

## 目標

- 對單一本機專案提供簡單、可預期的 CLI。
- 將正式規格、變更與任務以一般 Markdown 檔案保存，讓人可直接閱讀和用 Git 追蹤。
- 透過明確的任務完成狀態和封存檢查，降低遺漏規格更新的機會。
- 提供 macOS 可直接執行的單一 CLI；程式避免不必要的作業系統專屬行為。

## 不做的事

- 不提供 Server、帳號、遠端 workspace、同步或多人認領。
- 不提供桌面 UI、看板或瀏覽器介面。
- 不呼叫或內嵌任何 AI 模型；AI agent 透過一般檔案和 CLI 使用 Speclet。
- 不提供討論、依賴排序、worktree 管理、品質關卡、套件自動更新或外掛系統。
- 不在 MVP 管理 Git commit、branch 或 push。

## 方案比較與選擇

| 方案 | 優點 | 代價 |
| --- | --- | --- |
| Go（採用） | CLI 和檔案處理直接；可編譯成單一執行檔；跨平台發佈簡單 | 仍需建立 MVP 所需的 CLI 結構與檔案處理流程 |
| Rust | 型別和錯誤處理能力強 | 對這個小型 MVP 來說，初期建置和開發成本較高 |
| TypeScript | 原型與 Markdown/YAML 處理容易 | 使用者需有 Node.js 執行環境，單檔交付較不直接 |

本設計採用 Go，因為目前優先目標是快速做出可安裝、可使用的本機 CLI。

## 使用方式與資料佈局

Speclet 在目標專案根目錄建立並管理 `speclet/`：

```text
speclet/
  config.yaml
  specs/<capability>/spec.md
  changes/<change-name>/
    proposal.md
    tasks.md
    specs/<capability>/spec.md
  changes/archive/<YYYY-MM-DD>-<change-name>/
```

資料檔都是 UTF-8 Markdown 或 YAML。`specs/` 是目前正式規格；`changes/` 是尚未封存的工作；封存完成的資料移到 `changes/archive/`。MVP 不使用資料庫，也不在專案外建立隱藏狀態目錄。

除 `speclet init` 外，CLI 命令從目前工作目錄向上尋找最近的一個 `speclet/`，找到後以該目錄作為專案根。巢狀專案採最近者優先；一路找不到時報錯且不建立檔案。`speclet init [PATH]` 只在指定目錄（省略時為目前目錄）初始化，不向上搜尋。

## CLI 行為

MVP 對外提供以下命令。人類可讀文字輸出是唯一輸出格式；JSON 輸出不在此版範圍。

- `speclet init [PATH]`：在指定目錄（省略時為目前目錄）初始化 `speclet/` 基本結構。若已有有效的 `speclet/`，回報已初始化且不改寫內容；若資料夾存在但結構無法辨認，報錯且不覆寫。
- `speclet change new <name>`：建立變更目錄、`proposal.md`、`tasks.md` 和空的 delta 規格目錄。重複名稱拒絕建立。
- `speclet change list`：列出所有進行中變更及任務完成數；若沒有變更，明確回報空清單。
- `speclet change show <name>`：顯示變更提案、任務和規格差異的檔案位置；找不到變更時以非零狀態結束。
- `speclet task done <change> <task-id>`：將指定任務由未完成改為完成。
- `speclet task undone <change> <task-id>`：將指定任務改回未完成。
- `speclet archive <name>`：先檢查任務與所有規格差異，再將差異套用到正式規格並移動變更目錄至封存區。檢查失敗時不移動變更，也不改寫正式規格。

CLI 的錯誤訊息寫到 stderr，成功摘要寫到 stdout；成功結束碼為 0，輸入錯誤、找不到目標或資料驗證失敗使用非零結束碼。精確的錯誤文案不屬於契約。

## 規格與變更格式

Speclet 只解析明確標記的需求標題、情境標題與平面任務 checkbox。其他普通 Markdown（例如說明文字或表格）保留原文，不參與封存操作。程式碼圍欄內看似需求或任務標題的文字視為一般內容。巢狀任務 checkbox 不屬於 MVP 支援格式；若出現在需要解析的任務清單，Speclet 拒絕修改或封存並指出位置，不把它靜默計入完成狀態。

正式規格以 capability 分檔。每項需求以唯一名稱識別：

```markdown
## Requirement: <name>

需求本文。

### Scenario: <name>
- **WHEN** <條件>
- **THEN** <可觀察結果>
```

變更中的 delta 檔使用 `## ADDED`、`## MODIFIED`、`## REMOVED` 三種區段。其下需求標題為 `### Requirement: <name>`，情境標題為 `#### Scenario: <name>`，形成清楚的 Markdown 階層。規則如下：

- `ADDED` 的需求名稱在正式規格中必須不存在。
- `MODIFIED` 與 `REMOVED` 的需求名稱在正式規格中必須存在且只能有一份。
- `archive` 先驗證全部 delta；有未知區段、重複名稱、缺少目標需求或其他驗證錯誤時，整次操作拒絕，指出檔案與原因。
- 驗證成功後，`ADDED` 新增需求、`MODIFIED` 以 delta 內容取代同名需求，並將 `#### Scenario` 正規化為正式規格中的 `### Scenario`；`REMOVED` 刪除同名需求。需求以外的普通 Markdown 和程式碼圍欄內容保留原文。
- 規格更新完成後才移動變更資料夾到封存區。已存在同名封存目錄時拒絕，不能覆蓋舊資料。

`tasks.md` 使用平面 Markdown checkbox，任務需有穩定 ID，例如 `- [ ] T001 建立 CLI 初始化命令`。CLI 只修改指定 ID 的 checkbox，不改任務文字或其他任務。不存在、重複或巢狀的任務 ID 都拒絕。

## 封存規則

封存前必須滿足：

1. 所有任務均已勾選；空任務清單視為完成。
2. 變更中的每一份 delta 都符合上述格式與名稱規則。
3. 所有將要寫入的正式規格路徑可建立或更新，且目標封存目錄不存在。

任何輸入或驗證錯誤都要保留進行中變更與正式規格的原狀。MVP 不自動修正文句或合併 Git 衝突；格式或內容衝突由使用者修正後重試。多份正式規格同時更新時，實作者須先驗證所有輸入，再處理暫存、替換與失敗復原，避免一般錯誤留下半套結果。

## 驗收條件

- 在空專案執行 `speclet init` 後，建立完整資料佈局；再次執行不破壞既有檔案。
- 除 init 外，在專案子目錄執行命令可找到最近的 `speclet/`；巢狀專案使用最近者。找不到時不建立資料。
- 建立變更後，`change list` 與 `change show` 能呈現正確名稱和任務狀態。
- 對有效任務 ID 執行 done/undone 只改該任務；錯誤、重複或巢狀 ID 不改檔。
- 封存只修改標記的需求區塊，保留前言、表格、程式碼圍欄等普通 Markdown 原文。
- 所有任務完成且 delta 有效時，archive 更新正式規格並封存變更。
- 任務未完成、delta 不合法、目標需求不存在/重複，或封存目錄已存在時，archive 以非零結束，保留正式規格和進行中變更。
- 程式碼圍欄中的 `Requirement`、`Scenario` 或 checkbox 範例不會被當成規格或任務。

## 尚留給實作者的選擇

- Go CLI 套件與目錄分層，保持單一可執行檔和易測試即可。
- YAML 欄位的最小集合，以及 `config.yaml` 是否在第一個可用版本只保留版本欄位。
- 終端機顏色、help 版面與精確文字；須支援非互動與 pipe 場景，不得讓 ANSI 色碼污染檔案內容。
- 多個正式規格檔同時更新時的暫存、替換與失敗復原細節；必須先完成驗證，避免一般輸入錯誤造成半套封存。
