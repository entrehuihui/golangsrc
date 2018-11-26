package mqtthandler

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
	"time"

	"github.com/brocaar/lorawan"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/maxiiot/humiture/common"
	"github.com/maxiiot/humiture/setting"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

type MQTTHandler struct {
	conn          mqtt.Client
	rxTopic       string
	mutex         sync.RWMutex
	updata        chan *DataUpPayloadChan
	applicationID int
	server        string
	username      string
	password      string
	tlsCert       string
}

// NewHandler create a new mqttHandler
func NewHandler(server, username, password, tlsCert string, applicationID int) (*MQTTHandler, error) {
	var rxTopic string
	if applicationID > 0 {
		rxTopic = fmt.Sprintf(setting.Cfg.MqttServer.RxTopic, strconv.Itoa(applicationID), "+")
	} else {
		rxTopic = fmt.Sprintf(setting.Cfg.MqttServer.RxTopic, "+", "+")
	}
	h := MQTTHandler{
		rxTopic:       rxTopic,
		applicationID: applicationID,
		updata:        make(chan *DataUpPayloadChan, 100),
		server:        server,
		username:      username,
		password:      password,
		tlsCert:       tlsCert,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(server)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetOnConnectHandler(h.onConnected)
	opts.SetConnectionLostHandler(h.onConnectionLost)
	opts.SetMaxReconnectInterval(time.Second * 30)

	if tlsCert != "" {
		tlsconfig, err := newTLSConfig(tlsCert)
		if err != nil {
			return nil, errors.Errorf("Error with the mqtt CA certificate: %s", err)
		} else {
			opts.SetTLSConfig(tlsconfig)
		}
	}

	log.WithField("server", server).Info("mqttHandler: connecting to mqtt broker")
	h.conn = mqtt.NewClient(opts)
	for {
		if token := h.conn.Connect(); token.Wait() && token.Error() != nil {
			log.Errorf("mqttHandler: connecting to broker error,will retry in 2s: %s", token.Error())
			time.Sleep(time.Second * 2)
		} else {
			break
		}
	}

	return &h, nil
}

func newTLSConfig(cafile string) (*tls.Config, error) {
	cert, err := ioutil.ReadFile(cafile)
	if err != nil {
		log.Errorf("mqttHandler: couldn't load cafile: %s", err)
		return nil, err
	}

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(cert)

	return &tls.Config{
		RootCAs: certpool,
	}, nil
}

// Close stops the handler
func (h *MQTTHandler) Close() error {
	log.Info("mqttHandler: closing handler")
	log.WithField("topic", h.rxTopic).Info("mqttHandler: unsubscribing from rx topic")
	if token := h.conn.Unsubscribe(h.rxTopic); token.Wait() && token.Error() != nil {
		return fmt.Errorf("mqttHandler: unsubscribe from %s error: %s", h.rxTopic, token.Error())
	}
	log.Info("mqttHandler: handing last items in queue")
	close(h.updata)
	return nil
}

func (h *MQTTHandler) onConnected(c mqtt.Client) {
	defer h.mutex.RUnlock()
	h.mutex.RLock()

	log.Info("mqttHandler: connected to mqtt brocker")
	for {
		fmt.Println("listent ", h.rxTopic)
		if token := h.conn.Subscribe(h.rxTopic, 2, h.rxPacketHandler); token.Wait() && token.Error() != nil {
			log.Errorf("mattHandler: subscribe error: %s", token.Error())
			time.Sleep(time.Second)
			continue
		}
		return
	}
}

func (h *MQTTHandler) rxPacketHandler(c mqtt.Client, msg mqtt.Message) {
	var rxPacket DataUpPayload
	if err := json.Unmarshal(msg.Payload(), &rxPacket); err != nil {
		log.Errorf("mqttHander: decode rx packet error: %s", err)
		return
	}
	// fmt.Println("Test", string(msg.Payload()))
	if data, err := hex.DecodeString(rxPacket.Data); err == nil &&
		data[0] == 0xff {
		log.WithFields(log.Fields{
			"devEUI":  rxPacket.DevEUI,
			"payload": rxPacket.Data,
			"length":  len(data),
		}).Debug("Original frame")

		// 原批量数据上报，打印调试信息
		// for i := 0; i < int(math.Ceil(float64(len(rxPacket.Data))/32.)); i++ {
		// 	end := (i + 1) * 32
		// 	if end > len(rxPacket.Data) {
		// 		end = len(rxPacket.Data)
		// 	}
		// 	log.WithFields(log.Fields{
		// 		"devEUI":  rxPacket.DevEUI,
		// 		"payload": rxPacket.Data[i*32 : end],
		// 	}).Debug("Original frame")
		// }

		if _, ok := common.SyncDevTime[rxPacket.DevEUI]; !ok {
			common.SyncDevTime[rxPacket.DevEUI] = false
		}
		// 设备请求同步时间
		if len(data) == 3 && data[1] == 0x00 && data[2] == 0xff {
			SyncTime(rxPacket.DevEUI)
			common.SyncDevTime[rxPacket.DevEUI] = false
		} else {
			// A类同步时间处理
			if isSync, ok := common.SyncDevTime[rxPacket.DevEUI]; ok && isSync {
				SyncTime(rxPacket.DevEUI)
				common.SyncDevTime[rxPacket.DevEUI] = false
			}
			dataChan := &DataUpPayloadChan{
				Data:    data,
				DevEUI:  rxPacket.DevEUI.String(),
				DevName: rxPacket.DeviceName,
			}
			h.updata <- dataChan
		}
	}
}

// 当MQTT连接错误，尝试重新连接
func (h *MQTTHandler) onConnectionLost(c mqtt.Client, reason error) {
	log.Errorf("mqttHandler: mqtt connection error: %s", reason)
}

// 处理上行数据
func (h *MQTTHandler) HandleRXPackets(ws map[*websocket.Conn]struct{}) {
	for rxPacket := range h.updata {
		go handleHumitures(rxPacket, ws)
	}
}

func (b *MQTTHandler) publish(devEUI lorawan.EUI64, v DataDownPayload) error {
	var txTopic string
	if b.applicationID > 0 {
		txTopic = fmt.Sprintf("application/%d/node/%s/tx", b.applicationID, devEUI)
	} else {
		return errors.New("publish error,must provide application id")
	}

	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"topic": txTopic,
		"qos":   0,
	}).Info("backend: publishing packet")
	if token := b.conn.Publish(txTopic, 0, false, bytes); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

var PubChan chan PublishChan

func (b *MQTTHandler) PublishData() {
	for pub := range PubChan {
		if err := b.publish(pub.DevEUI, pub.Payload); err != nil {
			log.Error("publish time error:", err)
		}
	}
}

// DataDownPayload represents a data-down payload.
type DataDownPayload struct {
	ApplicationID int64         `json:"applicationID,string"`
	DevEUI        lorawan.EUI64 `json:"devEUI"`
	Confirmed     bool          `json:"confirmed"`
	FPort         uint8         `json:"fPort"`
	Data          []byte        `json:"data"`
	//Object        json.RawMessage `json:"object,ommitempty"`
}

type PublishChan struct {
	DevEUI  lorawan.EUI64
	Payload DataDownPayload
}

// DownlinkTime hex foramt:
// header code year month day hour minute second tail
//   1     1    2     1    1   1     1       1     1
func SyncTime(devEUI lorawan.EUI64) {
	//  取东八区时间
	now := time.Now().UTC().Add(time.Hour * 8)
	buf := bytes.NewBuffer([]byte{})
	buf.Write([]byte{0xff, 0x10})
	year := uint16(now.Year())
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, year)
	buf.Write(b)
	month := byte(now.Month())
	buf.WriteByte(month)
	day := byte(now.Day())
	buf.WriteByte(day)
	hour := byte(now.Hour())
	buf.WriteByte(hour)
	minute := byte(now.Minute())
	buf.WriteByte(minute)
	second := byte(now.Second())
	buf.WriteByte(second)
	buf.WriteByte(0xff)
	log.WithField("devEUI", devEUI).Debugf("downlink:%x", buf.Bytes())
	//now_bs64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	PubChan <- PublishChan{
		DevEUI: devEUI,
		Payload: DataDownPayload{
			DevEUI:    devEUI,
			Confirmed: false,
			FPort:     10,
			Data:      buf.Bytes(),
		},
	}
}
