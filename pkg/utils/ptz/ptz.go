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

// 获取 direction 方向型
/**
 *
 * @param options
 *        type:
 *        speed:default 30
 *        index:
 * @returns {string}
 */
func PTZCmd(ptzType PTZ_Type, speed uint8, index uint8) string {
	if speed == 0 {
		speed = 50
	}
	var value4, vlaue5, value6, value7 byte
	// 第四个字节。
	value4 = ptzMap[ptzType]

	switch ptzType {
	case Up:
	case Down:
		// 字节6 垂直控制速度相对值
		value6 = byte(speed)
		// 字节7 地址高四位ob0000_0000
		// indexValue6 = 0x00;
	case ApertureFar:
	case ApertureNear:
		// 字节6 光圈速度
		value6 = byte(speed)
	case Right:
	case Left:
		// 字节5 水平控制速度相对值
		vlaue5 = byte(speed)
	case FocusFar:
		// 字节5 聚焦速度
		vlaue5 = byte(speed)
	case FocusNear:
		// 字节5 聚焦速度
		vlaue5 = byte(speed)
	case LeftUp:
	case LeftDown:
	case RightUp:
	case RightDown:
		// 字节5 水平控制速度相对值
		vlaue5 = byte(speed)
		// 字节6 垂直控制速度相对值
		value6 = byte(speed)
	case ZoomFar:
		value7 = byte(speed)
	case ZoomNear:
		// 字节7 镜头变倍控制速度相对值 zoom
		value7 = byte(speed)
	case CalPos:
		value6 = byte(index)
	case DelPos:
		value6 = byte(index)
	case SetPos:
		value6 = byte(index)
	case WiperClose:
	case WiperOpen:
		vlaue5 = 0x01
	default:
	}
	return ptzCmdToString(value4, vlaue5, value6, value7)
}

func ptzCmdToString(value4 byte, value5 byte, value6 byte, value7 byte) string {
	cmd := make([]byte, 8)
	// 首字节以05H开头
	cmd[0] = 0xA5
	// 组合码，高4位为版本信息v1.0,版本信息0H，低四位为校验码
	cmd[1] = 0x0F
	// 地址的低8位
	cmd[2] = 0x01

	// 指令码
	if value4 != 0 {
		cmd[3] = value4
	}
	//  数 据 1
	if value5 != 0 {
		cmd[4] = value5
	}
	//  数 据 2
	if value6 != 0 {
		cmd[5] = value6
	}
	// 组合码 2 ,高 4 位是数据 3,低 4 位是地址的高 4 位;在后续叙述中,没有特别指明的高 4 位 ,表示该4位与所指定的功能无关
	if value7 != 0 {
		cmd[6] = value7
	}

	// 校验码, 为前面的第 1 ~ 7 字节的算术和的低 8 位,即算术和对 256 取模后的结果 。
	// 字节 8 = (字节 1 + 字节 2 + 字节 3 + 字节 4 + 字节 5 + 字节 6 + 字节 7)%256。

	i := 0
	for _, v := range cmd {
		i += int(v)
	}
	cmd[7] = byte(i % 256)
	return strings.ToUpper(hex.EncodeToString(cmd))
}
