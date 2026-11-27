# CLI 版本資訊實作計畫

> **執行方式：** 使用 `superpowers:executing-plans` 逐項執行本計畫；每個行為以測試先行。

**目標：** 讓 Speclet CLI 在 help 與 `--version` 顯示可辨識的應用程式版本。

**架構：** 新增 `internal/version`，正式建置以 linker flag 注入 Git tag；一般 Go 建置從 `debug.ReadBuildInfo` 取得 VCS revision 並顯示 `dev+<短 SHA>`。執行中的 CLI 不呼叫 Git，若 build info 沒有 revision 則顯示 `dev`。

**技術：** Go 標準函式庫、`go test`、`go build -ldflags`。

**規格：** `speclet/changes/cli-version-info/specs/cli-version/spec.md`；依據 [GitHub issue #10](https://github.com/chiisen/Speclet/issues/10)。

## 方法與取捨

- 正式版本透過 `-ldflags -X` 注入精確 Git tag，開發版本從 Go build info 取得 revision。安裝後不依賴 Git，也能保留來源資訊；採用此法。
- 執行時呼叫 Git 查詢 tag／commit，無法在一般安裝環境、非 repo 目錄或缺少 Git 時穩定工作。
- 在程式碼寫死版本字串，會隨開發與發版過期，且不能區分開發建置。

## 全域限制

- 正式版本使用語意化 Git tag，例如 `v0.1.0`；快速演進期間維持 `v0.x.y`。
- 未注入正式 tag 的開發建置顯示 `dev+` 與短 commit SHA；build info 無 revision 時顯示 `dev`。
- `--version` 與 help 均顯示應用程式版本；Go 版本與 `config.yaml` 格式版本各自維持原意。
- 版本查詢在專案初始化或語系設定錯誤時仍可執行。

## 審查重點

- linker flag 不存在或值為空：應回到可辨識的 `dev`，不 panic。
- VCS revision 短於七字元：不得切片越界。
- `--version` 不應讀寫專案資料，也不應依賴目前目錄或 Git 可執行檔。
- help 必須包含版本，原有命令與語系文字維持可見。
- 正式建置命令必須拒絕沒有精確 tag 的 HEAD，避免無意中產生錯誤 release 版本。

---

### 任務 1：建立 CLI 版本規格與測試

**檔案：**
- 新增：`speclet/changes/cli-version-info/proposal.md`
- 新增：`speclet/changes/cli-version-info/tasks.md`
- 新增：`speclet/changes/cli-version-info/specs/cli-version/spec.md`
- 新增：`internal/version/version_test.go`
- 修改：`internal/cli/run_test.go`

**介面：**
- 規格提供正式 tag、開發 SHA、版本旗標及 help 的驗收行為。
- 測試預期新增 `version.String()` 與 `speclet --version` 行為。

- [x] 建立 `cli-version-info` Speclet change，提案引用 issue #10，並新增 `cli-version` delta 規格。
- [x] 測試正式 tag 優先、開發版短 SHA、缺少 revision 的 fallback。
- [x] 測試 `--version` 精確輸出且 help 包含同一版本。
- [x] 執行相關 Go 測試，確認因版本實作缺失而失敗。

### 任務 2：實作版本解析及 CLI 輸出

**檔案：**
- 新增：`internal/version/version.go`
- 修改：`internal/cli/run.go`

**介面：**
- `internal/version.String() string` 回傳注入 tag、`dev+<7字元 SHA>` 或 `dev`。
- `Run` 處理 `--version`，help 開頭顯示 `version.String()`。

- [x] 以 linker-injected tag 優先，否則從 `runtime/debug.ReadBuildInfo` 的 `vcs.revision` 產生開發版字串。
- [x] 將版本分支置於語系設定錯誤檢查之前，維持版本查詢的獨立性。
- [x] 執行版本套件與 CLI 測試，確認全部通過。

### 任務 3：文件化開發與正式建置版本

**檔案：**
- 修改：`README.md`
- 修改：`CHANGELOG.md`

**介面：**
- README 提供標準安裝／開發建置及正式 tag 建置方式。
- 正式建置從 `git describe --tags --exact-match HEAD` 取得 tag，透過 `-ldflags` 傳入版本套件。

- [x] 說明 `--version`、help、`dev+短 SHA` 與正式 tag 的顯示規則。
- [x] 新增 Windows PowerShell 與 POSIX shell 正式 tag 建置範例，沒有精確 tag 時停止建置。
- [x] 以繁體中文更新 CHANGELOG。

### 任務 4：驗證並封存 Speclet change

**檔案：**
- 更新：`speclet/changes/cli-version-info/tasks.md`
- 更新：`speclet/specs/cli-version/spec.md`（由 archive 套用 delta）

- [x] 執行 `go test ./...`、`go build` 與實際 `--version`／help 驗收。
- [x] 使用 tag linker flag 建置正式版本並驗證輸出；以一般建置驗證開發 SHA。
- [x] 執行 `git diff --check`，確認 `speclet change show`、`speclet change list`。
- [x] 全部待辦完成且正式規格符合交付後封存 change。

