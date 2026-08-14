# Kiosk App (Wails + Vue)

一个类似信息亭的全屏桌面应用。启动时显示 loading 视频，后台通过 Go 后端检测目标网页是否可达（无 502/404 等），确认健康后再加载网页。

## 项目结构

```
kiosk-app/
├── main.go              # Wails 主入口，全屏无边框窗口
├── app.go               # Go 后端：URL 健康检查逻辑
├── wails.json           # Wails 项目配置
├── go.mod               # Go 模块定义
├── frontend/
│   ├── index.html       # 前端 HTML 入口
│   ├── package.json     # 前端依赖
│   ├── vite.config.js   # Vite 构建配置
│   └── src/
│       ├── main.js      # Vue 入口
│       ├── App.vue      # 主组件：loading 视频 + URL 检测 + iframe
│       └── assets/
│           └── loading.mp4  # 替换为你自己的 loading 视频
```

## 工作原理

1. 应用启动 → 全屏显示 loading 视频（循环播放）
2. Vue 组件挂载后，通过 Wails 绑定调用 Go 后端的 `CheckURLHealth()` 方法
3. Go 后端发起 HTTP GET 请求检测目标 URL：
   - 状态码 200-399 → 返回 "ok"
   - 状态码 4xx/5xx（如 404、502）→ 返回 "error: HTTP xxx"
   - 网络不可达 → 返回 "error: ..."
4. 如果检测不通过，每 5 秒自动重试，并在视频上叠加错误提示
5. 检测通过后，隐藏 loading 视频，显示 iframe 加载目标网页

## 配置

编辑 `frontend/src/App.vue` 顶部的常量：

```js
// 目标网页地址
const TARGET_URL = "https://www.example.com";

// Loading 视频路径（放在 frontend/src/assets/ 下）
const LOADING_VIDEO = "./assets/loading.mp4";

// 健康检查重试间隔（毫秒）
const RETRY_INTERVAL = 5000;
```

## 前置条件

- [Go](https://go.dev/dl/) 1.25+
- [Node.js](https://nodejs.org/) 18+
- [Wails v2](https://wails.io/docs/gettingstarted/installation) (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

## 构建与运行

```bash
# 1. 安装前端依赖
cd frontend && npm install

# 2. 放置 loading 视频
#    将你的 loading.mp4 复制到 frontend/src/assets/loading.mp4

# 3. 返回项目根目录，开发模式运行
cd ..
wails dev

# 4. 生产构建
wails build
```

构建产物在 `build/bin/` 目录下。

## 打包 Linux 包

### 方式一：GitHub Actions（推荐，无需本地 Linux 编译环境）

推送 `v*` 标签（如 `v1.0.0`）或手动触发 workflow，会自动打包出 Linux 的 **AMD64** 和 **ARM64** 两个架构的 AppImage：

```bash
git tag v1.0.0
git push origin v1.0.0
```

构建完成后，在 GitHub 仓库的 **Actions** 页面下载对应产物（Artifacts）。

### 方式二：本地 Linux 打包

在 Ubuntu / Debian 上安装编译依赖：

```bash
sudo apt-get update
sudo apt-get install -y \
  pkg-config \
  libgtk-3-dev \
  libwebkit2gtk-4.0-dev \
  libayatana-appindicator3-dev \
  libfuse2 \
  patchelf
```

安装 Wails CLI 并构建：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 本机架构（x86_64）
wails build --platform linux/amd64

# 或 ARM64 目标（需先配置交叉编译工具链）
wails build --platform linux/arm64
```

产物为 `build/bin/aigc-student`（可执行文件）及 `build/bin/aigc-student.AppImage`。

> AppImage 运行前通常需要 `--appimage-extract-and-run` 或安装 libfuse2，且需可执行权限：`chmod +x aigc-student.AppImage`。

## 说明

- 如果未提供 `loading.mp4`，前端会使用 CSS 动画作为 fallback loading 效果
- 窗口为全屏无边框模式，适合信息亭/嵌入式部署
- 按 `Alt+F4` 可退出应用