# ghproxy-go

[English](#english) | [中文](#中文)

## English

ghproxy-go is a high-performance reverse proxy for GitHub resources written in Go. It helps accelerate access to GitHub resources by proxying various GitHub URLs.

### Features

- Fast and lightweight, built with Go Fiber framework
- Supports multiple GitHub URL patterns
- Docker support
- Easy to deploy and configure

### Installation

#### Binary Release
Download the pre-compiled binary from the [Releases](https://github.com/nerdneilsfield/ghproxy-go/releases) page.

#### Docker
```bash
# Using Docker Hub
docker pull nerdneils/ghproxy-go

# Using GitHub Container Registry
docker pull ghcr.io/nerdneilsfield/ghproxy-go
```

### Usage

```bash
ghproxy-go [flags]
ghproxy-go [command]
```

#### Available Commands
- `completion`: Generate the autocompletion script for the specified shell
- `help`: Help about any command 
- `run`: Start the proxy server
- `version`: Show version information

#### Flags
- `-h, --help`: Help for ghproxy-go
- `-v, --verbose`: Enable verbose output

#### Run Command Flags
- `-H, --host string`: Host to listen on (default "0.0.0.0")
- `-P, --port int`: Port to listen on (default 8080) 
- `-J, --proxy-jsdelivr`: Enable jsdelivr proxy

### Supported URL Patterns

- GitHub Releases/Archives: `github.com/<author>/<repo>/releases/*` or `github.com/<author>/<repo>/archive/*`
- GitHub Blob/Raw: `github.com/<author>/<repo>/blob/*` or `github.com/<author>/<repo>/raw/*`
- GitHub Info/Git: `github.com/<author>/<repo>/info/*` or `github.com/<author>/<repo>/git-*`
- Raw Content: `raw.githubusercontent.com/<author>/<repo>/*`
- Gist: `gist.githubusercontent.com/<author>/*`
- GitHub Keys: `github.com/<author>.keys`

## 中文

ghproxy-go 是一个用 Go 语言编写的 GitHub 资源反向代理工具，通过代理各种 GitHub URL 来加速访问 GitHub 资源。

### 特性

- 基于 Go Fiber 框架，快速且轻量
- 支持多种 GitHub URL 模式
- 支持 Docker 部署
- 易于部署和配置

### 安装

#### 二进制安装
从 [Releases](https://github.com/nerdneilsfield/ghproxy-go/releases) 页面下载预编译的二进制文件。

#### Docker 安装
```bash
# 使用 Docker Hub
docker pull nerdneils/ghproxy-go

# 使用 GitHub Container Registry
docker pull ghcr.io/nerdneilsfield/ghproxy-go
```

### 使用方法

```bash
ghproxy-go [flags]
ghproxy-go [command]
```

#### 可用命令
- `completion`: 生成指定 shell 的自动补全脚本
- `help`: 显示帮助信息
- `run`: 启动代理服务器
- `version`: 显示版本信息

#### 全局参数
- `-h, --help`: 显示帮助信息
- `-v, --verbose`: 启用详细输出

#### Run 命令参数
- `-H, --host string`: 监听主机地址 (默认 "0.0.0.0")
- `-P, --port int`: 监听端口 (默认 8080)
- `-J, --proxy-jsdelivr`: 启用 jsdelivr 代理

### 支持的 URL 模式

- GitHub 发布/存档: `github.com/<作者>/<仓库>/releases/*` 或 `github.com/<作者>/<仓库>/archive/*`
- GitHub Blob/Raw: `github.com/<作者>/<仓库>/blob/*` 或 `github.com/<作者>/<仓库>/raw/*`
- GitHub Info/Git: `github.com/<作者>/<仓库>/info/*` 或 `github.com/<作者>/<仓库>/git-*`
- Raw 内容: `raw.githubusercontent.com/<作者>/<仓库>/*`
- Gist: `gist.githubusercontent.com/<作者>/*`
- GitHub Keys: `github.com/<作者>.keys`
