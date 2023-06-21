package sip

import (
	"fmt"
	"strings"
	"time"

	"github.com/panjjo/gosip/utils"
)

// DefaultProtocol DefaultProtocol
var DefaultProtocol = "udp"

var (
	// TcpProtocolKey tcp 的协议key
	TcpProtocolKey = "tcpConn"
	// UdpProtocolKey udp 的协议key
	UdpProtocolKey = "udpConn"
)

// DefaultSipVersion DefaultSipVersion
var DefaultSipVersion = "SIP/2.0"

// Port number
type Port uint16

// NewPort NewPort
func NewPort(port int) *Port {
	newPort := Port(port)
	return &newPort
}

// Clone clone
func (port *Port) Clone() *Port {
	if port == nil {
		return nil
	}
	newPort := *port
	return &newPort
}

func (port *Port) String() string {
	if port == nil {
		return ""
	}
	return fmt.Sprintf("%d", *port)
}

// Equals Equals
func (port *Port) Equals(other interface{}) bool {
	if p, ok := other.(*Port); ok {
		return Uint16PtrEq((*uint16)(port), (*uint16)(p))
	}

	return false
}

// MaybeString  wrapper
type MaybeString interface {
	String() string
	Equals(other interface{}) bool
}

// String string
type String struct {
	Str string
}

func (str String) String() string {
	return str.Str
}

// Equals Equals
func (str String) Equals(other interface{}) bool {
	if v, ok := other.(String); ok {
		return str.Str == v.Str
	}

	return false
}

// ContentTypeSDP SDP contenttype
var ContentTypeSDP = ContentType("application/sdp")

// ContentTypeXML XML contenttype
var ContentTypeXML = ContentType("Application/MANSCDP+xml")

var (
	// CatalogXML 获取设备列表xml样式<?xml version="1.0" encoding="GB2312"?>
	CatalogXML = `<?xml version="1.0" encoding="GB2312"?>
<Query>
  <CmdType>Catalog</CmdType>
  <SN>%d</SN>
  <DeviceID>%s</DeviceID>
</Query>
`
	// RecordInfoXML 获取录像文件列表xml样式<?xml version="1.0" encoding="GB2312"?>
	RecordInfoXML = `<?xml version="1.0"?>
<Query>
<CmdType>RecordInfo</CmdType>
<SN>%d</SN>
<DeviceID>%s</DeviceID>
<StartTime>%s</StartTime>
<EndTime>%s</EndTime>
<Secrecy>0</Secrecy>
<Type>time</Type>
</Query>
`
	// DeviceInfoXML 查询设备详情xml样式 <?xml version="1.0" encoding="GB2312"?>
	DeviceInfoXML = `<?xml version="1.0"?>
<Query>
<CmdType>DeviceInfo</CmdType>
<SN>%d</SN>
<DeviceID>%s</DeviceID>
</Query>
`
	//<! - 命令类型:设备配置查询(必选)->
	//<elementname="CmdType"fixed ="ConfigDownload"/> <! - - 命 令 序 列 号 (必 选 )- ->
	//<elementname="SN" type="integer"minInclusivevalue="1"/> <! - - 目 标 设 备 编 码 (必 选 )- ->
	//<elementname="DeviceID" type="tg:deviceIDType"/>
	//<! - 查询配置参数类型(必选),可查询的配置类型包括基本参数配置:BasicParam,视频参
	//数 范 围 :VideoParamOpt,SVAC 编 码 配 置 :SVACEncodeConfig,SVAC 解 码 配 置 :SVACDe- codeConfig。 可 同 时 查 询 多 个 配 置 类 型 ,各 类 型 以 “/”分 隔 ,可 返 回 与 查 询 SN 值 相 同 的 多 个 响
	//应 ,每 个 响 应 对 应 一 个 配 置 类 型 。- ->
	//<elementname="ConfigType" type="string"/>
	// ConfigDownloadInfoXML 查询设备配置查询xml样式 <?xml version="1.0" encoding="GB2312"?>
	ConfigDownloadInfoXML = `<?xml version="1.0"?>
<Query>
<CmdType>ConfigDownload</CmdType>
<SN>%d</SN>
<DeviceID>%s</DeviceID>
<ConfigType>BasicParam</ConfigType>
</Query>
`
	//移动设备位置数据查询
	//<! - 命令类型:移动设备位置数据查询(必选)-> <elementname="CmdType"fixed ="MobilePosition"/>
	//<! - - 命 令 序 列 号 (必 选 )- ->
	//<elementname="SN" type="integer"minInclusivevalue="1"/>
	//<! - - 查 询 移 动 设 备 编 码 (必 选 )- -> <elementname="DeviceID"type="tg:deviceIDType"/>
	//<! - 移动设备位置信息上报时间间隔,单位:秒,默认值5(可选)-> <elementname="Interval"type="integer"/>
	MobilePositionInfoXML = `<?xml version="1.0"?>
<Query>
<CmdType>MobilePosition</CmdType>
<SN>%d</SN>
<DeviceID>%s</DeviceID>
<Interval>5</Interval>
</Query>
`
)

// GetDeviceInfoXML 获取设备详情指令
func GetDeviceInfoXML(id string) []byte {
	return []byte(fmt.Sprintf(DeviceInfoXML, utils.RandInt(100000, 999999), id))
}

// GetCatalogXML 获取NVR下设备列表指令
func GetCatalogXML(id string) []byte {
	return []byte(fmt.Sprintf(CatalogXML, utils.RandInt(100000, 999999), id))
}

// GetRecordInfoXML 获取录像文件列表指令
func GetRecordInfoXML(id string, sceqNo int, start, end int64) []byte {
	return []byte(fmt.Sprintf(RecordInfoXML, sceqNo, id, time.Unix(start, 0).Format("2006-01-02T15:04:05"),
		time.Unix(end, 0).Format("2006-01-02T15:04:05")))
}

// GetConfigDownloadInfoXML 获取设备基本配置指令
func GetConfigDownloadInfoXML(id string) []byte {
	return []byte(fmt.Sprintf(ConfigDownloadInfoXML, utils.RandInt(100000, 999999), id))
}

// GetMobilePositionInfoXML 获取设备基本配置指令
func GetMobilePositionInfoXML(id string) []byte {
	return []byte(fmt.Sprintf(MobilePositionInfoXML, utils.RandInt(100000, 999999), id))
}

// RFC3261BranchMagicCookie RFC3261BranchMagicCookie
const RFC3261BranchMagicCookie = "z9hG4bK"

// GenerateBranch returns random unique branch ID.
func GenerateBranch() string {
	return strings.Join([]string{
		RFC3261BranchMagicCookie,
		utils.RandString(32),
	}, "")
}
