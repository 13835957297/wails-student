package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"encoding/json"
	"io"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.restartTeacher()
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
    teacherIP := "192.168.31.125"  // 固定教师端 IP
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