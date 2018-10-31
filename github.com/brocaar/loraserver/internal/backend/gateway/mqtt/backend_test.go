package mqtt

import (
	"os"
	"testing"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/loraserver/internal/backend"
	"github.com/brocaar/loraserver/internal/backend/gateway/marshaler"
	"github.com/brocaar/loraserver/internal/common"
	"github.com/brocaar/lorawan"
)

type BackendTestSuite struct {
	suite.Suite

	backend    backend.Gateway
	redisPool  *redis.Pool
	mqttClient paho.Client
}

func (ts *BackendTestSuite) SetupSuite() {
	assert := require.New(ts.T())
	log.SetLevel(log.ErrorLevel)

	redisURL := "redis://localhost:6379/1"
	if v := os.Getenv("TEST_REDIS_URL"); v != "" {
		redisURL = v
	}

	ts.redisPool = common.NewRedisPool(redisURL, 10, 0)

	mqttServer := "tcp://127.0.0.1:1883/1"
	var mqttUsername string
	var mqttPassword string

	if v := os.Getenv("TEST_MQTT_SERVER"); v != "" {
		mqttServer = v
	}
	if v := os.Getenv("TEST_MQTT_USERNAME"); v != "" {
		mqttUsername = v
	}
	if v := os.Getenv("TEST_MQTT_PASSWORD"); v != "" {
		mqttPassword = v
	}

	opts := paho.NewClientOptions().AddBroker(mqttServer).SetUsername(mqttUsername).SetPassword(mqttPassword)
	ts.mqttClient = paho.NewClient(opts)
	token := ts.mqttClient.Connect()
	token.Wait()
	assert.NoError(token.Error())

	var err error
	ts.backend, err = NewBackend(ts.redisPool, Config{
		Server:                mqttServer,
		Username:              mqttUsername,
		Password:              mqttPassword,
		CleanSession:          true,
		UplinkTopicTemplate:   "gateway/+/rx",
		DownlinkTopicTemplate: "gateway/{{ .MAC }}/tx",
		StatsTopicTemplate:    "gateway/+/stats",
		AckTopicTemplate:      "gateway/+/ack",
		ConfigTopicTemplate:   "gateway/{{ .MAC }}/config",
	})
	assert.NoError(err)

	ts.backend.(*Backend).setGatewayMarshaler(lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8}, marshaler.Protobuf)
}

func (ts *BackendTestSuite) TearDownSuite() {
	assert := require.New(ts.T())

	assert.NoError(ts.backend.Close())
}

func (ts *BackendTestSuite) SetupTest() {
	MustFlushRedis(ts.redisPool)
}

func (ts *BackendTestSuite) TestUplinkFrame() {
	assert := require.New(ts.T())

	uplinkFrame := gw.UplinkFrame{
		PhyPayload: []byte{1, 2, 3, 4},
		TxInfo: &gw.UplinkTXInfo{
			Frequency: 868100000,
		},
		RxInfo: &gw.UplinkRXInfo{
			GatewayId: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		},
	}

	b, err := proto.Marshal(&uplinkFrame)
	assert.NoError(err)
	uplinkFrame.XXX_sizecache = 0
	uplinkFrame.TxInfo.XXX_sizecache = 0
	uplinkFrame.RxInfo.XXX_sizecache = 0

	token := ts.mqttClient.Publish("gateway/0102030405060708/rx", 0, false, b)
	token.Wait()
	assert.NoError(token.Error())

	receivedUplink := <-ts.backend.RXPacketChan()
	assert.EqualValues(uplinkFrame, receivedUplink)
}

func (ts *BackendTestSuite) TestGatewayStats() {
	assert := require.New(ts.T())

	gatewayStats := gw.GatewayStats{
		GatewayId: []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}

	b, err := proto.Marshal(&gatewayStats)
	assert.NoError(err)
	gatewayStats.XXX_sizecache = 0

	token := ts.mqttClient.Publish("gateway/0102030405060708/stats", 0, false, b)
	token.Wait()
	assert.NoError(token.Error())

	receivedStats := <-ts.backend.StatsPacketChan()
	assert.EqualValues(gatewayStats, receivedStats)
}

func (ts *BackendTestSuite) TestDownlinkTXAck() {
	assert := require.New(ts.T())

	downlinkTXAck := gw.DownlinkTXAck{
		GatewayId: []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	b, err := proto.Marshal(&downlinkTXAck)
	assert.NoError(err)

	token := ts.mqttClient.Publish("gateway/0102030405060708/ack", 0, false, b)
	token.Wait()
	assert.NoError(token.Error())

	receivedAck := <-ts.backend.DownlinkTXAckChan()
	if !proto.Equal(&downlinkTXAck, &receivedAck) {
		assert.Equal(downlinkTXAck, receivedAck)
	}
}

func (ts *BackendTestSuite) TestSendDownlinkFrame() {
	assert := require.New(ts.T())

	downlinkFrameChan := make(chan gw.DownlinkFrame)
	token := ts.mqttClient.Subscribe("gateway/+/tx", 0, func(c paho.Client, msg paho.Message) {
		var pl gw.DownlinkFrame
		if err := proto.Unmarshal(msg.Payload(), &pl); err != nil {
			panic(err)
		}

		downlinkFrameChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	downlinkFrame := gw.DownlinkFrame{
		PhyPayload: []byte{1, 2, 3, 4},
		TxInfo: &gw.DownlinkTXInfo{
			GatewayId: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		},
	}
	assert.NoError(ts.backend.SendTXPacket(downlinkFrame))
	downlinkFrame.TxInfo.XXX_sizecache = 0
	downlinkFrame.XXX_sizecache = 0

	downlinkReceived := <-downlinkFrameChan
	assert.EqualValues(downlinkFrame, downlinkReceived)
}

func (ts *BackendTestSuite) TestSendGatewayConfiguration() {
	assert := require.New(ts.T())

	gatewayConfigChan := make(chan gw.GatewayConfiguration)
	token := ts.mqttClient.Subscribe("gateway/+/config", 0, func(c paho.Client, msg paho.Message) {
		var pl gw.GatewayConfiguration
		if err := proto.Unmarshal(msg.Payload(), &pl); err != nil {
			panic(err)
		}
		gatewayConfigChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	gatewayConfig := gw.GatewayConfiguration{
		GatewayId: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Version:   "1.2.3",
	}
	assert.NoError(ts.backend.SendGatewayConfigPacket(gatewayConfig))
	gatewayConfig.XXX_sizecache = 0

	configReceived := <-gatewayConfigChan
	assert.Equal(gatewayConfig, configReceived)
}

func TestBackend(t *testing.T) {
	suite.Run(t, new(BackendTestSuite))
}
