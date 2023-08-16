package gb28181

import (
	"context"
	"encoding/xml"
	"time"

	"github.com/go-av/gosip/pkg/server"
	"github.com/go-av/gosip/pkg/utils"
)

type Catalog struct {
	XMLName  xml.Name      `xml:"Response"`
	CmdType  CmdType       `xml:"CmdType"`
	SN       int64         `xml:"SN"`
	DeviceID string        `xml:"DeviceID"`
	SumNum   int           `xml:"SumNum"`
	Item     []CatalogItem `xml:"DeviceList>Item"`
}

type CatalogItem struct {

	/*
		deviceIDType: 设备编码类型
			<! -在取值为行政区划时可为2、4、6、8位,其他情况取值为20位。->
			<paternvalue="(\d{2}|\d{4}|\d{6}|\d{8}|\d{20})"/>
	*/

	DeviceID        string  `xml:"DeviceID" `                 // 设备/区域/系统编码(必选)
	Name            string  `xml:"Name"`                      // 设备/区域/系统名称(必选)
	Manufacturer    string  `xml:"Manufacturer"`              // 当为设备时,设备厂商(必选)
	Model           string  `xml:"Model"`                     // 当为设备时,设备型号(必选)
	Owner           string  `xml:"Owner"`                     // 当为设备时,设备归属(必选)
	CivilCode       string  `xml:"CivilCode"`                 // 行政区域 (必 选 )
	Block           string  `xml:"Block,omitempty"`           // 警 区 (可 选 )
	Address         string  `xml:"Address"`                   // 当为设备时,安装地址(必选)
	Parental        int     `xml:"Parental"`                  // 当为设备时,是否有子设备(必选)1有,0没有
	ParentID        string  `xml:"ParentID"`                  // 父设备/区域/系统ID(必选)
	SafetyWay       int     `xml:"SafetyWay"`                 // 信令安全模式(可选)缺省为0; 0:不采用;2:S/MIME签名方式;3:S/ MIME 加密签名同时采用方式;4:数字摘要方式
	RegisterWay     int     `xml:"RegisterWay"`               // 注 册 方 式 (必 选 )缺 省 为 1;1: 符 合 IETF RFC 3261 标 准 的 认 证 注 册 模 式 ;2 : 基 于 口 令 的 双 向 认 证 注 册 模 式 ;3 : 基 于 数 字 证 书 的 双 向 认 证 注 册 模 式 - -
	CertNum         string  `xml:"CertNum"`                   // 证书序列号(有证书的设备必选)
	Certifiable     int     `xml:"Certifiable"`               // 证书有效标识(有证书的设备必选)缺省为0;证书有效标识:0:无效 1: 有 效
	ErrCode         int     `xml:"ErrCode"`                   // 无效原因码(有证书且证书无效的设备必选)
	EndTime         string  `xml:"EndTime"`                   // 证书终止有效期(有证书的设备必选)
	Secrecy         int     `xml:"Secrecy"`                   // 保密属性(必选)缺省为0;0:不涉密,1:涉密
	IPAddres        string  `xml:"IPAddres,omitempty"`        // 设备 / 区域 / 系统 IP 地址 ( 可选 )
	Port            int     `xml:"Port,omitempty"`            // 设备/区域/系统端口(可选)
	Pasword         string  `xml:"Pasword,omitempty"`         // 设备口令(可选)
	Status          string  `xml:"Status"`                    // 设备状态 (必选)  ON OFF
	Longitude       float32 `xml:"Longitude,omitempty"`       // 经度 (可选 )
	Latitude        float32 `xml:"Latitude,omitempty"`        // 纬度 (可选 )
	PTZType         int     `xml:"PTZType,omitempty"`         // 摄像机类型扩展,标识摄像机类型:1-球机;2-半球;3-固定枪机;4-遥控枪 机。当目录项为摄像机时可选。
	PositionType    int     `xml:"PositionType,omitempty"`    // 摄像机位置类型扩展。1-省际检查站、2-党政机关、3-车站码头、4-中心广场 、5-体育场馆 、6-商业中心 、7-宗教场所 、8-校园周边 、9-治安复杂区域 、10-交通干线。当目录项为摄像机时可选。
	RoomType        int     `xml:"RoomType,omitempty"`        // 摄像机安装位置室外、室内属性。1-室外、2-室内。当目录项为摄像机时可 选 ,缺 省 为 1。
	UseType         int     `xml:"UseType,omitempty"`         // 摄像机用途属性。1-治安、2-交通、3-重点。当目录项为摄像机时可选。
	SupplyLightType int     `xml:"SupplyLightType,omitempty"` // 摄像机补光属性。1-无补光、2-红外补光、3-白光补光。当目录项为摄像机 时 可 选 ,缺 省 为 1。
	DirectionType   int     `xml:"DirectionType,omitempty"`   // 摄像机监视方位属性。1-东、2-西、3-南、4-北、5-东南、6-东北、7-西南、8-西 北。当目录项为摄像机时且为固定摄像机或设置看守位摄像机时可选。
	Resolution      string  `xml:"Resolution,omitempty"`      // 摄像机支持的分辨率,可有多个分辨率值,各个取值间以 / 分隔。分辨率 取值参见附录F中SDPf字段规定。当目录项为摄像机时可选。
}

/*
<? xmlversion="1.0"?>
<Response>
    <CmdType>Catalog </CmdType>
    <SN>17430</SN>
    <DeviceID>64010000001110000001</DeviceID>
    <SumNum>100</SumNum>
    <DeviceList Num=1>
        <Item>
            <DeviceID>64010000001330000001</DeviceID>
            <Name>Camera1</ Name>
            <Manufacturer>Manufacturer1</ Manufacturer>
            <Model>Model1</ Model>
            <Owner>Owner1</ Owner>
            <CivilCode>CivilCode1</ CivilCode>
            <Block > Block 1</ Block >
            <Address>Address 1</ Address>
            <Parental>1 </Parental>
            <ParentID>64010000001110000001 </ParentID>
            <SafetyWay>0 </SafetyWay>
            <RegisterWay>1</ RegisterWay>
            <CertNum >CertNum 1</ CertNum >
            <Certifiable>0</ Certifiable>
            <ErrCode>400</ ErrCode>
            <EndTime>2010-11-11T19:46:17 </EndTime>
            <Secrecy>0 </Secrecy>
            <IPAddres>192.168.3.81 </IPAddres>
            <Port>5060</ Port>
            <Pasword>Pasword1 </Pasword>
            <Status>Status1 </Status>
            <Longitude>171.3</Longitude>
            <Longitude > 34 .2 </Longitude>
        </Item>
    </DeviceList>
</Response>


*/

type catalogWithCache struct {
	catalog *Catalog
	timer   *time.Timer
}

func (g *GB28181) GetCatalog(client server.Client) (int64, error) {
	sn := time.Now().Unix()
	_, err := g.SendMessage(client, &Query{
		CmdType:  CmdType__Catalog,
		SN:       sn,
		DeviceID: client.User(),
	})

	if err != nil {
		return 0, err
	}

	return sn, nil
}

func (g *GB28181) Catalog(ctx context.Context, client server.Client, body []byte) (*server.Response, error) {
	cl := &Catalog{}
	if err := utils.XMLDecode(body, cl); err != nil {
		return nil, err
	}

	g.handler.Catalog(ctx, client, cl)
	return server.NewResponse(200, "success."), nil
}
