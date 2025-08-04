
```bash
   go install github.com/tc-hib/go-winres@latest
```


```bash
   export PATH=$PATH:$(go env GOPATH)/bin
```

运行脚本（在 macOS 上）：

```bash
   export RUNNER_OS="macOS" && export PLATFORM="darwin" && export ARCH="amd64" && export EXT="" && ./scripts/ci-build-client.sh
```

这会生成了 macOS 的 DMG 文件：dist/smart-finder-client-darwin-amd64.dmg。

如果需要在 Windows 上运行，只需要将环境变量改为：

```bash
export RUNNER_OS="Windows" && export PLATFORM="windows" && export ARCH="amd64" && export EXT=".exe"
```