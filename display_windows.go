package main

import (
	"syscall"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procEnumDisplayMonitors = user32.NewProc("EnumDisplayMonitors")
	procGetMonitorInfoW     = user32.NewProc("GetMonitorInfoW")
)

// Rect represents a Windows RECT structure
type Rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// MonitorInfoExW represents MONITORINFOEXW structure
type MonitorInfoExW struct {
	CbSize    uint32
	RcMonitor Rect
	RcWork    Rect
	DwFlags   uint32
	SzDevice  [32]uint16
}

const (
	MONITORINFOF_PRIMARY = 0x00000001
)

// DisplayInfo holds information about a detected display
type DisplayInfo struct {
	IsPrimary bool
	Left      int32
	Top       int32
	Width     int32
	Height    int32
}

// displays stores all enumerated displays; populated by EnumDisplayMonitors callback
var displays []DisplayInfo

func enumDisplayCallback(hMonitor syscall.Handle, hdc syscall.Handle, lprcMonitor *Rect, dwData uintptr) uintptr {
	var info MonitorInfoExW
	info.CbSize = uint32(unsafe.Sizeof(info))

	ret, _, _ := procGetMonitorInfoW.Call(
		uintptr(hMonitor),
		uintptr(unsafe.Pointer(&info)),
	)
	if ret == 0 {
		return 1 // continue enumeration
	}

	displays = append(displays, DisplayInfo{
		IsPrimary: (info.DwFlags & MONITORINFOF_PRIMARY) != 0,
		Left:      info.RcMonitor.Left,
		Top:       info.RcMonitor.Top,
		Width:     info.RcMonitor.Right - info.RcMonitor.Left,
		Height:    info.RcMonitor.Bottom - info.RcMonitor.Top,
	})

	return 1 // continue enumeration
}

// GetAllDisplays enumerates all monitors using Windows API and returns display info
func GetAllDisplays() []DisplayInfo {
	displays = nil

	procEnumDisplayMonitors.Call(
		0, // HDC = NULL
		0, // lprcClip = NULL
		syscall.NewCallback(enumDisplayCallback),
		0, // dwData
	)

	return displays
}