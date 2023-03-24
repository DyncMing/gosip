package sipapi

import (
	"fmt"
	"github.com/prometheus/common/log"
	"net/http"

	"github.com/panjjo/gosip/db"
	sip "github.com/panjjo/gosip/sip/s"
	"github.com/panjjo/gosip/utils"
	"github.com/sirupsen/logrus"
)

// MessageReceive 接收到的请求数据最外层，主要用来判断数据类型
type MessageReceive struct {
	CmdType string `xml:"CmdType"`
	SN      int    `xml:"SN"`
}

func handlerMessage(req *sip.Request, tx *sip.Transaction) {
	u, ok := parserDevicesFromRequest(req)
	if !ok {
		// 未解析出来源用户返回错误
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil))
		return
	}
	// 判断是否存在body数据
	if len, have := req.ContentLength(); !have || len.Equals(0) {
		// 不存在就直接返回的成功
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
		return
	}
	body := req.Body()
	message := &MessageReceive{}
	log.Debugln("request body: ", string(body))

	if err := utils.XMLDecode(body, message); err != nil {
		logrus.Warnln("Message Unmarshal xml err:", err, "body:", string(body))
		// 有些body xml发送过来的不带encoding ，而且格式不是utf8的，导致xml解析失败，此处使用gbk转utf8后再次尝试xml解析
		body, err = utils.GbkToUtf8(body)
		if err != nil {
			logrus.Errorln("message gbk to utf8 err", err)
		}
		if err := utils.XMLDecode(body, message); err != nil {
			logrus.Errorln("Message Unmarshal xml after gbktoutf8 err:", err, "body:", string(body))
			tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil))
			return
		}
	}
	switch message.CmdType {
	case "Catalog":
		// 设备列表
		sipMessageCatalog(u, body)
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
		return
	case "Keepalive":
		// heardbeat
		if err := sipMessageKeepalive(u, body); err == nil {
			tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
			// 心跳后同步注册设备列表信息
			sipCatalog(u)
			//同步设备基本信息
			//sipConfigDownload(u)
			//获取移动设备位置信息
			//sipMobilePosition(u)
			//查询设备信息
			//sipDeviceInfo(u)
			//CheckDeviceOffline()
			return
		} else {
			logrus.Errorln("Keepalive => err：", err)
		}
	case "RecordInfo":
		// 设备音视频文件列表
		sipMessageRecordInfo(u, body)
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
	case "DeviceInfo":
		// 主设备信息
		log.Debugln("DeviceInfo: ", string(body))
		sipMessageDeviceInfo(u, body)
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
		return
	case "ConfigDownload":
		log.Debugln("ConfigDownload: ", string(body))
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
	case "MobilePosition":
		log.Debugln("MobilePosition: ", string(body))
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
	case "DataTransfer":
		log.Debugln("DataTransfer: ", string(body))
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
	}
	tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil))
}

func handlerRegister(req *sip.Request, tx *sip.Transaction) {
	// 判断是否存在授权字段
	if hdrs := req.GetHeaders("Authorization"); len(hdrs) > 0 {
		fromUser, ok := parserDevicesFromRequest(req)
		if !ok {
			return
		}
		user := Devices{DeviceID: fromUser.DeviceID}
		if err := db.Get(db.DBClient, &user); err == nil {
			if !user.Regist {
				// 如果数据库里用户未激活，替换user数据
				fromUser.ID = user.ID
				fromUser.Name = user.Name
				fromUser.PWD = "123456" //user.PWD
				user = fromUser
			}
			user.addr = fromUser.addr
			authenticateHeader := hdrs[0].(*sip.GenericHeader)
			auth := sip.AuthFromValue(authenticateHeader.Contents)
			auth.SetPassword("123456" /*user.PWD*/)
			auth.SetUsername(user.DeviceID)
			auth.SetMethod(string(req.Method()))
			auth.SetURI(auth.Get("uri"))
			if auth.CalcResponse() == auth.Get("response") {
				// 验证成功
				// 记录活跃设备
				user.source = fromUser.source
				user.addr = fromUser.addr
				_activeDevices.Store(user.DeviceID, user)
				if !user.Regist {
					// 第一次激活，保存数据库
					user.Regist = true
					db.DBClient.Save(&user)
					logrus.Infoln("new user regist,id:", user.DeviceID)
				}
				tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
				//同步设备基本信息
				//sipDeviceBasicConfig(user)
				//获取移动设备位置信息
				//sipMobilePosition(user)
				// 注册成功后查询设备信息，获取制作厂商等信息
				//go notify(notifyDevicesRegister(user))
				sipDeviceInfo(fromUser)
				sipCatalog(fromUser)
				return
			}
		} else {
			logrus.Errorln("register user err: ", err)
			user.addr = fromUser.addr
			authenticateHeader := hdrs[0].(*sip.GenericHeader)
			auth := sip.AuthFromValue(authenticateHeader.Contents)
			auth.SetPassword("123456")
			auth.SetUsername(fromUser.DeviceID)
			auth.SetMethod(string(req.Method()))
			auth.SetURI(auth.Get("uri"))
			if auth.CalcResponse() == auth.Get("response") {
				// 验证成功
				// 记录活跃设备
				user.source = fromUser.source
				user.addr = fromUser.addr
				user.Model = fromUser.Model
				user.Name = fromUser.Model
				user.Firmware = fromUser.Firmware
				user.Manufacturer = fromUser.Manufacturer
				user.Host = fromUser.Host
				user.ID = fromUser.ID
				user.Region = fromUser.Region
				_activeDevices.Store(user.DeviceID, user)
				if !user.Regist {
					// 第一次激活，保存数据库
					user.Regist = true
					db.DBClient.Save(&user)
					logrus.Infoln("new user regist,id:", user.DeviceID)
				}
				tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
				// 注册成功后查询设备信息，获取制作厂商等信息
				//go notify(notifyDevicesRegister(user))
				go sipDeviceInfo(fromUser)
				return
			}
		}
	} else {
		logrus.Errorln("StatusUnauthorized: ", string(req.Body()), "Headers: ", req.GetHeaders("Authorization"))
	}

	resp := sip.NewResponseFromRequest("", req, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), nil)
	resp.AppendHeader(&sip.GenericHeader{HeaderName: "WWW-Authenticate", Contents: fmt.Sprintf("Digest nonce=\"%s\", algorithm=MD5, realm=\"%s\",qop=\"auth\"", utils.RandString(32), _sysinfo.Region)})
	tx.Respond(resp)
}

func handlerUnAuthorizationRegister(req *sip.Request, tx *sip.Transaction) {
	// 判断是否存在授权字段
	fromUser, ok := parserDevicesFromRequest(req)
	if !ok {
		return
	}
	user := Devices{DeviceID: fromUser.DeviceID}
	if err := db.Get(db.DBClient, &user); err == nil {
		if !user.Regist {
			// 如果数据库里用户未激活，替换user数据
			fromUser.ID = user.ID
			fromUser.Name = user.Name
			fromUser.PWD = "123456" //user.PWD
			user = fromUser
		}
		user.addr = fromUser.addr
		// 验证成功
		// 记录活跃设备
		user.source = fromUser.source
		user.addr = fromUser.addr
		_activeDevices.Store(user.DeviceID, user)
		if !user.Regist {
			// 第一次激活，保存数据库
			user.Regist = true
			db.DBClient.Save(&user)
			logrus.Infoln("new user regist,id:", user.DeviceID)
		}
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
		//同步设备基本信息
		//sipDeviceBasicConfig(user)
		//获取移动设备位置信息
		//sipMobilePosition(user)
		// 注册成功后查询设备信息，获取制作厂商等信息
		//go notify(notifyDevicesRegister(user))
		go sipDeviceInfo(fromUser)
		return
	} else {
		logrus.Errorln("register user err: ", err)
		user.addr = fromUser.addr
		// 验证成功
		// 记录活跃设备
		user.source = fromUser.source
		user.addr = fromUser.addr
		user.Model = fromUser.Model
		user.Name = fromUser.Model
		user.Firmware = fromUser.Firmware
		user.Manufacturer = fromUser.Manufacturer
		user.Host = fromUser.Host
		user.ID = fromUser.ID
		user.Region = fromUser.Region
		_activeDevices.Store(user.DeviceID, user)
		if !user.Regist {
			// 第一次激活，保存数据库
			user.Regist = true
			db.DBClient.Save(&user)
			logrus.Infoln("new user regist,id:", user.DeviceID)
		}
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
		// 注册成功后查询设备信息，获取制作厂商等信息
		//go notify(notifyDevicesRegister(user))
		go sipDeviceInfo(fromUser)
		return
	}
	resp := sip.NewResponseFromRequest("", req, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), nil)
	resp.AppendHeader(&sip.GenericHeader{HeaderName: "WWW-Authenticate", Contents: fmt.Sprintf("Digest nonce=\"%s\", algorithm=MD5, realm=\"%s\",qop=\"auth\"", utils.RandString(32), _sysinfo.Region)})
	tx.Respond(resp)
}

func handlerSubscribe(req *sip.Request, tx *sip.Transaction) {
	u, ok := parserDevicesFromRequest(req)
	if !ok {
		// 未解析出来源用户返回错误
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil))
		return
	}
	// 判断是否存在body数据
	if len, have := req.ContentLength(); !have || len.Equals(0) {
		// 不存在就直接返回的成功
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
		return
	}
	body := req.Body()
	message := &MessageReceive{}
	logrus.Debugln("handlerSubscribe: ", string(body))

	if err := utils.XMLDecode(body, message); err != nil {
		logrus.Warnln("Message Unmarshal xml err:", err, "body:", string(body))
		// 有些body xml发送过来的不带encoding ，而且格式不是utf8的，导致xml解析失败，此处使用gbk转utf8后再次尝试xml解析
		body, err = utils.GbkToUtf8(body)
		if err != nil {
			logrus.Errorln("message gbk to utf8 err", err)
		}
		if err := utils.XMLDecode(body, message); err != nil {
			logrus.Errorln("Message Unmarshal xml after gbktoutf8 err:", err, "body:", string(body))
			tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil))
			return
		}
	}
	switch message.CmdType {
	case "Catalog":
		// 设备列表
		sipMessageCatalog(u, body)
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
		return
	case "Keepalive":
		// heardbeat
		if err := sipMessageKeepalive(u, body); err == nil {
			tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
			// 心跳后同步注册设备列表信息
			//sipCatalog(u)
			//CheckDeviceOffline()
			//同步设备基本信息
			//sipConfigDownload(u)
			//获取移动设备位置信息
			//sipMobilePosition(u)
			//查询设备信息
			//sipDeviceInfo(u)
			return
		}
	case "RecordInfo":
		// 设备音视频文件列表
		log.Debug("handlerSubscribe RecordInfo=: ", string(body))
		sipMessageRecordInfo(u, body)
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
	case "DeviceInfo":
		// 主设备信息
		log.Debug("handlerSubscribe DeviceInfo: ", string(body))
		sipMessageDeviceInfo(u, body)
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
		return
	case "ConfigDownload":
		log.Debug("handlerSubscribe ConfigDownload: ", string(body))
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
	case "MobilePosition":
		log.Debug("handlerSubscribe MobilePosition: ", string(body))
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
	case "DataTransfer":
		log.Debug("handlerSubscribe DataTransfer: ", string(body))
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
	}
	tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil))
}
