package main

import (
	"embed"
	"context"
	"fmt"
	"net/http"
	"time"

	"encoding/json"
	"io"

	"log"
	"net"
	"os"
	"path/filepath"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed frontend/dist/loading.webm
var videoAssets embed.FS

type App struct {
	ctx context.Context
	videoServerURL string
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.restartTeacher()
	go a.StartVideoServer()
}

func (a *App) domReady(ctx context.Context) {
	// displays := GetAllDisplays()

	// // Find the target display:
	// // - If there's a non-primary display (external big screen via HDMI splitter), use it
	// // - Otherwise fall back to the primary display
	// var target *DisplayInfo
	// for i := range displays {
	// 	d := &displays[i]
	// 	if d.IsPrimary {
	// 		target = d
	// 		break
	// 	}
	// 	 fmt.Printf("屏幕 %d: IsPrimary=%v, Left=%d, Top=%d, Width=%d, Height=%d\n",
    //     i, d.IsPrimary, d.Left, d.Top, d.Width, d.Height)
	// }
	// if target == nil && len(displays) > 0 {
	// 	target = &displays[0]
	// }

	// if target != nil {
	// 	// Move window to the target display first, then go fullscreen
	// 	wailsRuntime.WindowSetPosition(ctx, int(target.Left), int(target.Top))
	// 	wailsRuntime.WindowSetSize(ctx, int(target.Width), int(target.Height))
	// }

	wailsRuntime.WindowSetAlwaysOnTop(ctx, true)
	wailsRuntime.WindowFullscreen(ctx)
	js := `(function(){
		var id = '__wails_hide_scrollbar__';
		if (document.getElementById(id)) return;
		var style = document.createElement('style');
		style.id = id;
		style.textContent =
			"*::-webkit-scrollbar{display:none!important;width:0!important;height:0!important;}" +
			"*{scrollbar-width:none!important;}" +
			"html,body{-ms-overflow-style:none!important;}";
		(document.head || document.documentElement).appendChild(style);
		document.addEventListener('contextmenu', e => e.preventDefault(), true);
	})();`
	wailsRuntime.WindowExecJS(ctx, js)
}

// CheckURLHealth checks if a given URL is reachable and returns a non-error HTTP status.
// It follows up to 3 redirects and treats any 2xx/3xx as healthy.
func (a *App) CheckURLHealth(url string) string {
	if url == "" {
		return "error: empty URL"
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(a.ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return "ok"
	}
	return fmt.Sprintf("error: HTTP %d", resp.StatusCode)
}


func (a *App) restartTeacher() {
    teacherIP := "192.168.20.200"  // 固定教师端 IP
    port := "9527"

    // 第一步：先 ping 确认教师端在线
    pingURL := fmt.Sprintf("http://%s:%s/ping", teacherIP, port)
    client := &http.Client{Timeout: 3 * time.Second}

    resp, err := client.Get(pingURL)
    if err != nil {
        fmt.Println("教师端不在线:", err)
        return
    }
    body, _ := io.ReadAll(resp.Body)
    resp.Body.Close()

    var result map[string]string
    json.Unmarshal(body, &result)
    if result["role"] != "teacher" {
        fmt.Println("目标不是教师端")
        return
    }

    // 第二步：发送重启指令
    restartURL := fmt.Sprintf("http://%s:%s/restart", teacherIP, port)
    _, err = client.Get(restartURL)
    if err != nil {
        fmt.Println("发送重启指令失败:", err)
        return
    }
    fmt.Println("已向教师端发送重启指令")
}

// GetTeacherStatus 前端可调用，查看教师端状态
// func (a *App) GetTeacherStatus() string {
//     teacherIP := "192.168.1.100"
//     port := "9527"
//     pingURL := fmt.Sprintf("http://%s:%s/ping", teacherIP, port)
//     client := &http.Client{Timeout: 2 * time.Second}

//     resp, err := client.Get(pingURL)
//     if err != nil {
//         return "offline"
//     }
//     resp.Body.Close()
//     return "online"
// }

// 启动本地视频文件服务器
// StartVideoServer 增强版：支持 Windows/Linux，支持 dev/production
func (a *App) StartVideoServer() string {
    if a.videoServerURL != "" {
        return a.videoServerURL
    }

    // 从嵌入的 assets 读取视频
    data, err := videoAssets.ReadFile("frontend/dist/loading.webm")
    if err != nil {
        log.Println("❌ 读取视频失败:", err)
        return ""
    }

    // 创建临时目录并写入视频
    tmpDir, err := os.MkdirTemp("", "wails-video-*")
    if err != nil {
        log.Println("创建临时目录失败:", err)
        return ""
    }

    videoPath := filepath.Join(tmpDir, "loading.webm")
    if err := os.WriteFile(videoPath, data, 0644); err != nil {
        os.RemoveAll(tmpDir)
        log.Println("写入视频失败:", err)
        return ""
    }

    // 启动 HTTP 服务器
    fs := http.FileServer(http.Dir(tmpDir))
    listener, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil {
        os.RemoveAll(tmpDir)
        log.Println("启动服务器失败:", err)
        return ""
    }

    a.videoServerURL = fmt.Sprintf("http://%s", listener.Addr().String())
    go func() {
        if err := http.Serve(listener, fs); err != nil {
            log.Println("服务器异常:", err)
        }
    }()

    log.Println("🚀 视频服务器启动于:", a.videoServerURL)
    return a.videoServerURL
}
func (a *App) GetVideoServerURL() string {
	return a.videoServerURL
}