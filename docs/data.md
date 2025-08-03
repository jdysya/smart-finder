## 数据存储路径

```go
case "darwin":
    appDataPath = filepath.Join(home, "Library", "Application Support", "Smart Finder")
case "windows":
    appDataPath = filepath.Join(home, "AppData", "Roaming", "Smart Finder")
```

存储路径在用户目录下的 `Library/Application Support/Smart Finder` 或 `AppData/Roaming/Smart Finder`。