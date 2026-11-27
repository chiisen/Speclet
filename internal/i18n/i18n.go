package i18n

import (
	"errors"
	"fmt"
	"strings"
)

const (
	English       = "en"
	TraditionalZH = "zh-TW"
)

type Language struct {
	Code        string
	EnglishName string
	NativeName  string
	IsDefault   bool
}

var languages = []Language{
	{Code: English, EnglishName: "English", NativeName: "英文", IsDefault: true},
	{Code: TraditionalZH, EnglishName: "Traditional Chinese", NativeName: "繁體中文"},
}

// SupportedLanguages returns the locales for which this binary includes a catalog.
func SupportedLanguages() []Language {
	return append([]Language(nil), languages...)
}

func IsSupported(code string) bool {
	for _, language := range languages {
		if language.Code == code {
			return true
		}
	}
	return false
}

func Name(code, displayLanguage string) string {
	for _, language := range languages {
		if language.Code == code {
			if displayLanguage == TraditionalZH {
				return language.NativeName
			}
			return language.EnglishName
		}
	}
	return code
}

var traditionalChinese = map[string]string{
	"missing command":                                      "缺少命令",
	"usage: speclet init [PATH]":                           "用法：speclet init [PATH]",
	"inspect speclet directory":                            "檢查 speclet 目錄失敗",
	"speclet init":                                         "初始化 Speclet 專案失敗",
	"Speclet project already initialized.":                 "Speclet 專案已初始化。",
	"Initialized Speclet project.":                         "已初始化 Speclet 專案。",
	"usage: speclet archive <name>":                        "用法：speclet archive <name>",
	"archive change":                                       "封存變更失敗",
	"Archived change %q.":                                  "已封存變更 %q。",
	"unknown command: %s":                                  "未知命令：%s",
	"usage: speclet task <done|undone> <change> <task-id>": "用法：speclet task <done|undone> <change> <task-id>",
	"find change":                                          "找不到變更",
	"update task":                                          "更新任務狀態失敗",
	"Marked task %s %s for change %q.":                     "已將變更 %q 的任務 %s 標記為%s。",
	"done":                                                 "完成",
	"undone":                                               "未完成",
	"usage: speclet change <new|list|show>":                "用法：speclet change <new|list|show>",
	"usage: speclet change new <name>":                     "用法：speclet change new <name>",
	"create change":                                        "建立變更失敗",
	"Created change %q.":                                   "已建立變更 %q。",
	"usage: speclet change list [--names-only | --json]":   "用法：speclet change list [--names-only | --json]",
	"use either --json or --names-only, not both; usage: speclet change list [--names-only | --json]": "--json 和 --names-only 只能擇一；用法：speclet change list [--names-only | --json]",
	"list changes":                          "列出變更失敗",
	"No active changes.":                    "目前沒有進行中的變更。",
	"%s (%d/%d tasks done)":                 "%s（%d/%d 項任務已完成）",
	"usage: speclet change show <name>":     "用法：speclet change show <name>",
	"show change":                           "顯示變更失敗",
	"Change: %s\nProposal: %s\nTasks: %s\n": "變更：%s\n提案：%s\n任務：%s\n",
	"Delta specs: none":                     "規格差異：無",
	"Delta specs:":                          "規格差異：",
	"unknown change command: %s":            "未知的變更命令：%s",
	"Speclet - local-first specification workflow\n\nUsage:\n  speclet init [PATH]\n  speclet change new <name>\n  speclet change list [--names-only | --json]\n  speclet change show <name>\n  speclet task done <change> <task-id>\n  speclet task undone <change> <task-id>\n  speclet archive <name>\n  speclet config language list\n  speclet config language show\n  speclet config language set <code>\n  speclet help\n\nUse --names-only to print one active change name per line without task counts.\nUse --json to print active change summaries as a JSON array.\nChanges use Markdown proposal.md, tasks.md, and specs/<capability>/spec.md files.\nArchive requires every task to be complete and every delta to be valid.\n": "Speclet - 本機優先的規格工作流程\n\n用法：\n  speclet init [PATH]\n  speclet change new <name>\n  speclet change list [--names-only | --json]\n  speclet change show <name>\n  speclet task done <change> <task-id>\n  speclet task undone <change> <task-id>\n  speclet archive <name>\n  speclet config language list\n  speclet config language show\n  speclet config language set <code>\n  speclet help\n\n--names-only 每行只輸出一個進行中的變更名稱，不附任務進度。\n--json 以 JSON 陣列輸出進行中的變更名稱和任務完成數。\n變更內容使用 Markdown 檔案：proposal.md、tasks.md 和 specs/<capability>/spec.md。\n封存前必須完成所有任務，且規格差異格式有效。\n",
	"usage: speclet config language <list|show|set> [code]":              "用法：speclet config language <list|show|set> [code]",
	"usage: speclet config language set <code>":                          "用法：speclet config language set <code>",
	"Supported languages:":                                               "支援的語系：",
	"Current language: %s (%s, default).":                                "目前語系：%s（%s，預設值）。",
	"Current language: %s (%s).":                                         "目前語系：%s（%s）。",
	"Language preference saved: %s (%s).":                                "已儲存語系偏好：%s（%s）。",
	"English":                                                            "English",
	"Traditional Chinese":                                                "繁體中文",
	"unsupported language %q; supported languages: %s":                   "不支援語系 %q；可選語系：%s",
	"user config directory is unavailable: %w":                           "無法取得使用者設定目錄：%w",
	"user settings path is not a regular file: %s":                       "使用者設定路徑不是一般檔案：%s",
	"read user settings %s: %w":                                          "讀取使用者設定 %s 失敗：%w",
	"parse user settings: %w":                                            "解析使用者設定失敗：%w",
	"language setting is empty":                                          "語系設定不可為空",
	"create user config directory: %w":                                   "建立使用者設定目錄失敗：%w",
	"create temporary settings file: %w":                                 "建立暫存設定檔失敗：%w",
	"write settings file: %w":                                            "寫入設定檔失敗：%w",
	"sync settings file: %w":                                             "同步設定檔失敗：%w",
	"close settings file: %w":                                            "關閉設定檔失敗：%w",
	"install settings file: %w":                                          "安裝設定檔失敗：%w",
	"missing settings file path":                                         "缺少設定檔路徑",
	"project path is empty":                                              "專案路徑不可為空",
	"resolve project path: %w":                                           "解析專案路徑失敗：%w",
	"create project directory: %w":                                       "建立專案目錄失敗：%w",
	"invalid existing speclet layout: %w":                                "既有 Speclet 目錄結構無效：%w",
	"inspect speclet directory: %w":                                      "檢查 Speclet 目錄失敗：%w",
	"prepare speclet layout: %w":                                         "準備 Speclet 目錄結構失敗：%w",
	"create %s: %w":                                                      "建立 %s 失敗：%w",
	"write config.yaml: %w":                                              "寫入 config.yaml 失敗：%w",
	"install speclet layout: %w":                                         "安裝 Speclet 目錄結構失敗：%w",
	"working directory is empty":                                         "工作目錄不可為空",
	"resolve working directory: %w":                                      "解析工作目錄失敗：%w",
	"invalid speclet layout at %s: %w":                                   "%s 的 Speclet 目錄結構無效：%w",
	"inspect %s: %w":                                                     "檢查 %s 失敗：%w",
	"speclet project not found from %s":                                  "從 %s 找不到 Speclet 專案",
	"speclet is not a directory":                                         "speclet 不是目錄",
	"config.yaml: %w":                                                    "讀取 config.yaml 失敗：%w",
	"config.yaml is not a regular file":                                  "config.yaml 不是一般檔案",
	"read config.yaml: %w":                                               "讀取 config.yaml 失敗：%w",
	"config.yaml is not valid UTF-8":                                     "config.yaml 不是有效的 UTF-8",
	"config.yaml must declare version: 1":                                "config.yaml 必須宣告 version: 1",
	"%s: %w":                                                             "%s：%w",
	"%s is not a directory":                                              "%s 不是目錄",
	"invalid speclet layout: %w":                                         "Speclet 目錄結構無效：%w",
	"another Speclet write is in progress, or a stale lock exists at %s": "另一個 Speclet 寫入程序正在執行，或 %s 留有舊鎖定",
	"acquire Speclet write lock: %w":                                     "取得 Speclet 寫入鎖定失敗：%w",
	"release Speclet write lock: %w":                                     "釋放 Speclet 寫入鎖定失敗：%w",
	"speclet root: %w":                                                   "Speclet 根目錄：%w",
	"read tasks.md: %w":                                                  "讀取 tasks.md 失敗：%w",
	"validate tasks.md: %w":                                              "驗證 tasks.md 失敗：%w",
	"change %q has unfinished tasks (%d/%d done)":                        "變更 %q 尚有未完成任務（%d/%d 項已完成）",
	"changes directory: %w":                                              "changes 目錄：%w",
	"changes directory":                                                  "changes 目錄",
	"archive directory: %w":                                              "archive 目錄：%w",
	"archive directory":                                                  "archive 目錄",
	"archive destination already exists: %s":                             "封存目的地已存在：%s",
	"inspect archive destination: %w":                                    "檢查封存目的地失敗：%w",
	"formal specs directory: %w":                                         "正式規格目錄：%w",
	"delta path escapes change directory: %s":                            "規格差異路徑超出 change 目錄：%s",
	"delta path must be specs/<capability>/spec.md: %s":                  "規格差異路徑必須是 specs/<capability>/spec.md：%s",
	"formal spec capability %q: %w":                                      "正式規格功能 %q：%w",
	"inspect formal spec capability %q: %w":                              "檢查正式規格功能 %q 失敗：%w",
	"multiple deltas target %s":                                          "多個規格差異都指向 %s",
	"read delta %s: %w":                                                  "讀取規格差異 %s 失敗：%w",
	"validate delta %s: %w":                                              "驗證規格差異 %s 失敗：%w",
	"formal spec target is not a regular file: %s":                       "正式規格目標不是一般檔案：%s",
	"read formal spec %s: %w":                                            "讀取正式規格 %s 失敗：%w",
	"inspect formal spec %s: %w":                                         "檢查正式規格 %s 失敗：%w",
	"apply delta %s: %w":                                                 "套用規格差異 %s 失敗：%w",
	"create formal spec directory %s: %w":                                "建立正式規格目錄 %s 失敗：%w",
	"inspect formal spec directory %s: %w":                               "檢查正式規格目錄 %s 失敗：%w",
	"stage formal spec %s: %w":                                           "準備正式規格 %s 失敗：%w",
	"write staged formal spec %s: %w":                                    "寫入暫存正式規格 %s 失敗：%w",
	"set staged formal spec mode: %w":                                    "設定暫存正式規格權限失敗：%w",
	"close staged formal spec: %w":                                       "關閉暫存正式規格失敗：%w",
	"replace formal spec %s: %w":                                         "取代正式規格 %s 失敗：%w",
	"rollback failed: %s":                                                "回復失敗：%s",
	"%v; %w":                                                             "%v；%w",
	"not a regular directory":                                            "不是一般目錄",
	"change %q already exists":                                           "變更 %q 已存在",
	"inspect change %q: %w":                                              "檢查變更 %q 失敗：%w",
	"prepare change %q: %w":                                              "準備變更 %q 失敗：%w",
	"set change directory mode: %w":                                      "設定變更目錄權限失敗：%w",
	"create change specs: %w":                                            "建立變更規格目錄失敗：%w",
	"create proposal.md: %w":                                             "建立 proposal.md 失敗：%w",
	"create tasks.md: %w":                                                "建立 tasks.md 失敗：%w",
	"publish change %q: %w":                                              "建立變更 %q 失敗：%w",
	"read changes directory: %w":                                         "讀取 changes 目錄失敗：%w",
	"change %q: %w":                                                      "變更 %q：%w",
	"change %q tasks.md: %w":                                             "變更 %q 的 tasks.md：%w",
	"read %s: %w":                                                        "讀取 %s 失敗：%w",
	"change %q specs: %w":                                                "變更 %q 的 specs 目錄：%w",
	"symbolic links are not supported: %s":                               "不支援符號連結：%s",
	"read change specs: %w":                                              "讀取變更規格失敗：%w",
	"invalid change name %q":                                             "變更名稱 %q 無效",
	"invalid change name %q: control characters are not allowed":         "變更名稱 %q 無效：不可包含控制字元",
	"not a regular file":                                                 "不是一般檔案",
	"delta is not valid UTF-8":                                           "規格差異不是有效的 UTF-8",
	"duplicate delta Requirement %q":                                     "規格差異中有重複的 Requirement：%q",
	"line %d: delta Requirement heading must use level 3":                "第 %d 行：delta 的 Requirement 標題必須使用三級標題",
	"line %d: delta Scenario heading must use level 4":                   "第 %d 行：delta 的 Scenario 標題必須使用四級標題",
	"line %d: %s name is empty":                                          "第 %d 行：%s 名稱不可為空",
	"line %d: unknown delta section %q":                                  "第 %d 行：未知的 delta 區段 %q",
	"line %d: Requirement is outside a delta section":                    "第 %d 行：Requirement 位於 delta 區段之外",
	"line %d: Scenario is outside a Requirement":                         "第 %d 行：Scenario 位於 Requirement 之外",
	"unclosed code fence":                                                "程式碼區塊未關閉",
	"parse formal spec: %w":                                              "解析正式規格失敗：%w",
	"parse delta: %w":                                                    "解析規格差異失敗：%w",
	"cannot add Requirement %q: it already exists":                       "無法新增 Requirement %q：該需求已存在",
	"cannot %s Requirement %q: it does not exist":                        "無法%s Requirement %q：該需求不存在",
	"delta produces invalid formal spec: %w":                             "規格差異產生的正式規格無效：%w",
	"formal spec is not valid UTF-8":                                     "正式規格不是有效的 UTF-8",
	"line %d: Requirement heading must use level 2":                      "第 %d 行：Requirement 標題必須使用二級標題",
	"line %d: Scenario heading must use level 3":                         "第 %d 行：Scenario 標題必須使用三級標題",
	"line %d: duplicate Requirement %q":                                  "第 %d 行：Requirement %q 重複",
	"tasks document is not valid UTF-8":                                  "任務文件不是有效的 UTF-8",
	"nested checkbox at line %d is unsupported":                          "第 %d 行不支援巢狀核取方塊",
	"task checkbox at line %d has no task ID":                            "第 %d 行的任務核取方塊缺少 task ID",
	"task ID %q is duplicated at line %d":                                "task ID %q 在第 %d 行重複",
	"invalid task ID %q":                                                 "task ID %q 無效",
	"inspect tasks file: %w":                                             "檢查任務檔案失敗：%w",
	"tasks file is not a regular file":                                   "任務檔案不是一般檔案",
	"inspect tasks directory: %w":                                        "檢查任務目錄失敗：%w",
	"tasks directory is not a regular directory":                         "任務目錄不是一般目錄",
	"read tasks file: %w":                                                "讀取任務檔案失敗：%w",
	"task ID %q is duplicated":                                           "task ID %q 重複",
	"task ID %q was not found":                                           "找不到 task ID %q",
	"stage tasks file: %w":                                               "準備暫存任務檔案失敗：%w",
	"write staged tasks file: %w":                                        "寫入暫存任務檔案失敗：%w",
	"set staged tasks file mode: %w":                                     "設定暫存任務檔案權限失敗：%w",
	"sync staged tasks file: %w":                                         "同步暫存任務檔案失敗：%w",
	"close staged tasks file: %w":                                        "關閉暫存任務檔案失敗：%w",
	"replace tasks file: %w":                                             "取代任務檔案失敗：%w",
}

type MessageError struct {
	format string
	args   []any
	cause  error
}

func Errorf(format string, args ...any) error {
	var cause error
	if strings.Contains(format, "%w") {
		for i := len(args) - 1; i >= 0; i-- {
			if err, ok := args[i].(error); ok {
				cause = err
				break
			}
		}
	}
	return &MessageError{format: format, args: append([]any(nil), args...), cause: cause}
}

func (e *MessageError) Error() string {
	return Text(English, e.format, e.args...)
}

func (e *MessageError) Unwrap() error {
	return e.cause
}

func Text(language, key string, args ...any) string {
	template := key
	if language == TraditionalZH {
		if translated, ok := traditionalChinese[key]; ok {
			template = translated
		}
	}
	if len(args) == 0 {
		return template
	}
	template = strings.ReplaceAll(template, "%w", "%v")
	values := append([]any(nil), args...)
	for i, value := range values {
		if err, ok := value.(error); ok {
			values[i] = ErrorText(language, err)
		}
	}
	return fmt.Sprintf(template, values...)
}

func ErrorText(language string, err error) string {
	if err == nil {
		return ""
	}
	var message *MessageError
	if errors.As(err, &message) {
		return Text(language, message.format, message.args...)
	}
	return err.Error()
}
