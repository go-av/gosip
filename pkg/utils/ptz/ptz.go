package ptz

import (
	"encoding/hex"
	"strings"
)

type PTZ_Type string

const (
	Stop         PTZ_Type = "stop"         // 停止
	Right        PTZ_Type = "right"        // 右
	Left         PTZ_Type = "left"         // 左
	Up           PTZ_Type = "up"           // 上
	Down         PTZ_Type = "down"         // 下
	LeftUp       PTZ_Type = "leftUp"       // 左上
	LeftDown     PTZ_Type = "leftDown"     // 左下
	RightUp      PTZ_Type = "rightUp"      // 右上
	RightDown    PTZ_Type = "rightDown"    // 右下
	ZoomFar      PTZ_Type = "zoomFar"      // 镜头 放大
	ZoomNear     PTZ_Type = "zoomNear"     // 镜头 缩小
	ApertureFar  PTZ_Type = "apertureFar"  // 光圈 缩小
	ApertureNear PTZ_Type = "apertureNear" // 光圈 放大
	FocusFar     PTZ_Type = "focusFar"     // 聚焦 近
	FocusNear    PTZ_Type = "focusNear"    // 聚焦 远
	SetPos       PTZ_Type = "setPos"       // 设置预设点
	CalPos       PTZ_Type = "calPos"       // 调用预设点
	DelPos       PTZ_Type = "delPos"       // 删除预设点
	WiperOpen    PTZ_Type = "wiperOpen"    // 雨刷开
	WiperClose   PTZ_Type = "wiperClose"   // 雨刷关
)

var ptzMap = map[PTZ_Type]byte{
	Stop: 0x00,

	Right: 0x01, // 0000 0001
	Left:  0x02, // 0000 0010
	Up:    0x08, // 0000 1000
	Down:  0x04, // 0000 0100

	LeftUp:    0x0A, // 0000 1010
	LeftDown:  0x06, // 0000 0110
	RightUp:   0x09, // 0000 1001
	RightDown: 0x05, // 0000 0101

	ZoomFar:  0x10, // 镜头 放大
	ZoomNear: 0x20, // 镜头 缩小

	ApertureFar:  0x48, // 光圈 缩小
	ApertureNear: 0x44, // 光圈 放大

	FocusFar:  0x42, // 聚焦 近
	FocusNear: 0x41, // 聚焦 远

	SetPos: 0x81, // 设置预设点
	CalPos: 0x82, // 调用预设点
	DelPos: 0x83, // 删除预设点

	WiperOpen:  0x8C, // 雨刷开
	WiperClose: 0x8D, // 雨刷关
}

// 0x19 :00011001
// 0x32 :00110010
// 0x4b :01001011
// 0x64 :01100100
// 0xFA :11111010
// 速度范围： 为00H~FFH
var SPEED_ARRAY = []byte{0x19, 0x32, 0x4b, 0x64, 0x7d, 0x96, 0xAF, 0xC8, 0xE1, 0xFA}

// 0x01 :000000001
// 0x02 :000000010
// 0x03 :000000011
// 0x10 :000010000
// 预置位范围：01H~FFH
var POSITION_ARRAY = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10}

// 0H~FH; 低四位地址是高四位0000
// 1,3,5,7,9,10,a,c,d,f
var ZOOM_ARRAY = []byte{0x10, 0x30, 0x50, 0x70, 0x90, 0xA0, 0xB0, 0xC0, 0xd0, 0xe0}

// 获取 direction 方向型
/**
 *
 * @param options
 *        type:
 *        speed:default 5
 *        index:
 * @returns {string}
 */
func PTZCmd(opt PTZ_Type, speed int, posindex int) string {
	var ptzSpeed = getPTZSpeed(speed)
	var indexValue3, indexValue4, indexValue5, indexValue6 byte
	// 第四个字节。
	indexValue3 = ptzMap[opt]

	switch opt {
	case Up:
	case Down:
		// 字节6 垂直控制速度相对值
		indexValue5 = ptzSpeed
		// 字节7 地址高四位ob0000_0000
		// indexValue6 = 0x00;
	case ApertureFar:
	case ApertureNear:
		// 字节6 光圈速度
		indexValue5 = ptzSpeed
		// 字节7 地址高四位ob0000_0000
		// indexValue6 = 0x00;
	case Right:
	case Left:
		// 字节5 水平控制速度相对值
		indexValue4 = ptzSpeed
		// 字节7 地址高四位ob0000_0000
		// indexValue6 = 0x00;
	case FocusFar:
	case FocusNear:
		// 字节5 聚焦速度
		indexValue4 = ptzSpeed
		// 字节7 地址高四位ob0000_0000
		// indexValue6 = 0x00;
	case LeftUp:
	case LeftDown:
	case RightUp:
	case RightDown:
		// 字节5 水平控制速度相对值
		indexValue4 = ptzSpeed
		// 字节6 垂直控制速度相对值
		indexValue5 = ptzSpeed
		// 字节7 地址高四位ob0000_0000
		// indexValue6 = 0x00;
	case ZoomFar:
		indexValue6 = getZoomSpeed(speed)
	case ZoomNear:
		// 字节7 镜头变倍控制速度相对值 zoom
		indexValue6 = getZoomSpeed(speed)
	case CalPos:
		indexValue5 = getPTZPositionIndex(posindex)
	case DelPos:
		indexValue5 = getPTZPositionIndex(posindex)
	case SetPos:
		// 第五个字节 00H
		// indexValue4 = 0x00;
		// 字节6 01H~FFH 位置。
		indexValue5 = getPTZPositionIndex(posindex)
	case WiperClose:
	case WiperOpen:
		// 字节5为辅助开关编号,取值为“1”表示雨刷控制。
		indexValue4 = 0x01
	default:
	}
	return PtzCmdToString(indexValue3, indexValue4, indexValue5, indexValue6)
}

func getPTZSpeed(speed int) byte {
	if speed <= 0 {
		speed = 5
	}

	return SPEED_ARRAY[speed-1]
}

func getZoomSpeed(speed int) byte {
	if speed == 0 {
		speed = 5
	}
	return ZOOM_ARRAY[speed-1]
}

func getPTZPositionIndex(index int) byte {
	return POSITION_ARRAY[index-1]
}

func PtzCmdToString(indexValue3 byte, indexValue4 byte, indexValue5 byte, indexValue6 byte) string {
	//
	cmd := make([]byte, 8)
	// 首字节以05H开头
	cmd[0] = 0xA5
	// 组合码，高4位为版本信息v1.0,版本信息0H，低四位为校验码
	cmd[1] = 0x0F
	// 校验码 = (cmd[0]的高4位+cmd[0]的低4位+cmd[1]的高4位)%16
	cmd[2] = 0x01
	//
	if indexValue3 != 0 {
		cmd[3] = indexValue3
	}
	if indexValue4 != 0 {
		cmd[4] = indexValue4
	}
	if indexValue5 != 0 {
		cmd[5] = indexValue5
	}
	if indexValue6 != 0 {
		cmd[6] = indexValue6
	}

	i := 0
	for _, v := range cmd {
		i += int(v)
	}

	cmd[7] = byte(i % 256)
	return strings.ToUpper(hex.EncodeToString(cmd))
}
