# 用 Speclet 管理任務狀態

以下假設專案已經執行過 `speclet init`。

## 建立一項任務

建立一個 change：

```sh
speclet change new greeting
```

在 `speclet/changes/greeting/tasks.md` 加入一項待辦：

```markdown
## 待辦事項

- [ ] T001 印出歡迎訊息
```

執行 `speclet change list`，會看到任務尚未完成：

```text
greeting (0/1 tasks done)
```

## 完成後更新狀態

T001 完成後，執行以下命令更新狀態：

```sh
speclet task done greeting T001
speclet change list
```

清單會顯示 `greeting (1/1 tasks done)`，而 `tasks.md` 中的任務也會改成 `- [x]`。

如果不小心標成完成，可用 `speclet task undone greeting T001` 改回未完成狀態。
