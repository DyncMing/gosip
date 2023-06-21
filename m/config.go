package m

import (
	"github.com/panjjo/gosip/utils"
	"net"
	"os"
	"strings"
	"time"

	"github.com/panjjo/gosip/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config Config
type Config struct {
	MOD       string            `json:"mod" yaml:"mod" mapstructure:"mod"`
	DB        db.Config         `json:"database" yaml:"database" mapstructure:"database"`
	LogLevel  string            `json:"logger" yaml:"logger" mapstructure:"logger"`
	Port      string            `json:"port" yaml:"port" mapstructure:"port"`
	Host      string            `json:"host" yaml:"host" mapstructure:"host"`
	API       string            `json:"api" yaml:"api" mapstructure:"api"`
	Secret    string            `json:"secret" yaml:"secret" mapstructure:"secret"`
	Media     MediaServer       `json:"media" yaml:"media" mapstructure:"media"`
	Stream    Stream            `json:"stream" yaml:"stream" mapstructure:"stream"`
	Record    RecordCfg         `json:"record" yaml:"record" mapstructure:"record"`
	GB28181   *SysInfo          `json:"gb28181" yaml:"gb28181" mapstructure:"gb28181"`
	Notify    map[string]string `json:"notify" yaml:"notify" mapstructure:"notify"`
	NotifyMap map[string]string
}

func (config Config) GetHostUri() string {
	return config.Host + ":" + config.Port
}

type RecordCfg struct {
	FilePath  string `json:"filepath" yaml:"filepath" mapstructure:"filepath"`
	Expire    int    `json:"expire" yaml:"expire"  mapstructure:"expire"`
	Recordmax int    `json:"recordmax" yaml:"recordmax"  mapstructure:"recordmax"`
}

// Stream Stream
type Stream struct {
	HLS  bool `json:"hls" yaml:"hls" mapstructure:"hls"`
	RTMP bool `json:"rtmp" yaml:"rtmp" mapstructure:"rtmp"`
}

// MediaServer MediaServer
type MediaServer struct {
	RESTFUL string `json:"restful" yaml:"restful" mapstructure:"restful"`
	HTTP    string `json:"http" yaml:"http" mapstructure:"http"`
	WS      string `json:"ws" yaml:"ws" mapstructure:"ws"`
	RTMP    string `json:"rtmp" yaml:"rtmp" mapstructure:"rtmp"`
	RTSP    string `json:"rtsp" yaml:"rtsp" mapstructure:"rtsp"`
	RTP     string `json:"rtp" yaml:"rtp" mapstructure:"rtp"`
	Secret  string `json:"secret" yaml:"secret" mapstructure:"secret"`
}

type SysInfo struct {
	db.DBModel
	// Region 当前域
	Region string `json:"region"   yaml:"region" mapstructure:"region"`
	// CID 通道id固定头部
	CID string `json:"cid"   yaml:"cid" mapstructure:"cid"`
	// CNUM 当前通道数
	CNUM int `json:"cnum" bson:"unum" yaml:"unum" mapstructure:"unum"`
	// DID 设备id固定头部
	DID string `json:"did" bson:"did" yaml:"did" mapstructure:"did"`
	// DNUM 当前设备数
	DNUM int `json:"dnum" bson:"dnum" yaml:"dnum" mapstructure:"dnum"`
	// LID 当前服务id
	LID         string `json:"lid" bson:"lid" yaml:"lid" mapstructure:"lid"`
	MediaServer bool
	// 媒体服务器接流地址
	MediaServerRtpIP net.IP `gorm:"-" json:"-"`
	// 媒体服务器接流端口
	MediaServerRtpPort int `gorm:"-"  json:"-"`
}

func DefaultInfo() *SysInfo {
	return MConfig.GB28181
}

var MConfig *Config

func LoadConfig() {
	viper.SetConfigType("yml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.SetDefault("logger", "debug")
	viper.SetDefault("udp", "0.0.0.0:5060")
	viper.SetDefault("api", "0.0.0.0:8090")
	viper.SetDefault("mod", "release")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatalln("init config error:", err)
	}
	logrus.Infoln("init config ok")
	MConfig = &Config{}
	err = viper.Unmarshal(&MConfig)
	if err != nil {
		logrus.Fatalln("init config unmarshal error:", err)
	}
	logrus.Infof("config :%+v", MConfig)
	level, _ := logrus.ParseLevel(MConfig.LogLevel)
	logrus.SetLevel(level)
	logrus.SetOutput(os.Stdout)
	//Log file segmentation hook
	hook := utils.NewLfsHook(time.Duration(7)*time.Hour, 60, "./logs", "gosip")
	logrus.AddHook(hook)
	db.DBClient, err = db.Open(MConfig.DB)
	if err != nil {
		logrus.Fatalln("init db error:", err)
	}
	db.DBClient.DB().SetMaxIdleConns(25)
	db.DBClient.DB().SetMaxOpenConns(100)

	db.DBClient.SetNowFuncOverride(func() interface{} {
		return time.Now().Unix()
	})
	//db.DBClient.LogMode(true)
	go db.KeepLive(db.DBClient, time.Minute)

	MConfig.MOD = strings.ToUpper(MConfig.MOD)
	notifyMap := map[string]string{}
	if MConfig.Notify != nil {
		for k, v := range MConfig.Notify {
			if v != "" {
				notifyMap[strings.ReplaceAll(k, "_", ".")] = v
			}
		}
	}
	MConfig.NotifyMap = notifyMap
	if MConfig.Record.Expire == 0 {
		MConfig.Record.Expire = 7
	}
	if MConfig.Record.Recordmax == 0 {
		MConfig.Record.Expire = 600
	}
}
