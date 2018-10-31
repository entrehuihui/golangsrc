package testsuite

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/loraserver/internal/config"
	"github.com/brocaar/loraserver/internal/gps"
	"github.com/brocaar/loraserver/internal/helpers"
	"github.com/brocaar/loraserver/internal/storage"
	"github.com/brocaar/loraserver/internal/uplink"
	"github.com/brocaar/lorawan"
)

type ClassBTestSuite struct {
	IntegrationTestSuite
}

func (ts *ClassBTestSuite) SetupSuite() {
	ts.DatabaseTestSuiteBase.SetupSuite()

	config.C.NetworkServer.NetworkSettings.ClassB.PingSlotDR = 2
	config.C.NetworkServer.NetworkSettings.ClassB.PingSlotFrequency = 868300000

	ts.CreateGateway(storage.Gateway{
		GatewayID: [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		Location: storage.GPSPoint{
			Latitude:  1.1234,
			Longitude: 1.1235,
		},
		Altitude: 10.5,
	})

	ts.CreateServiceProfile(storage.ServiceProfile{
		AddGWMetadata: true,
	})

	// device-profile
	ts.CreateDeviceProfile(storage.DeviceProfile{
		SupportsClassB: true,
	})

	// device
	ts.CreateDevice(storage.Device{
		DevEUI: lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	})
}

func (ts *ClassBTestSuite) TestUplink() {
	assert := require.New(ts.T())

	queueItems := []storage.DeviceQueueItem{
		{
			DevEUI:     ts.Device.DevEUI,
			FRMPayload: []byte{1, 2, 3, 4},
			FPort:      1,
			FCnt:       1,
		},
		{
			DevEUI:     ts.Device.DevEUI,
			FRMPayload: []byte{1, 2, 3, 4},
			FPort:      1,
			FCnt:       2,
		},
		{
			DevEUI:     ts.Device.DevEUI,
			FRMPayload: []byte{1, 2, 3, 4},
			FPort:      1,
			FCnt:       3,
		},
	}
	for i := range queueItems {
		assert.NoError(storage.CreateDeviceQueueItem(ts.DB(), &queueItems[i]))
	}

	// device-session
	ds := storage.DeviceSession{
		DeviceProfileID:  ts.Device.DeviceProfileID,
		ServiceProfileID: ts.Device.ServiceProfileID,
		RoutingProfileID: ts.Device.RoutingProfileID,
		DevEUI:           ts.Device.DevEUI,
		JoinEUI:          lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},

		DevAddr:               lorawan.DevAddr{1, 2, 3, 4},
		FNwkSIntKey:           [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		SNwkSIntKey:           [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		NwkSEncKey:            [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		FCntUp:                8,
		NFCntDown:             5,
		EnabledUplinkChannels: []int{0, 1, 2},
		PingSlotNb:            1,
		RX2Frequency:          869525000,
	}

	now := time.Now().UTC().Truncate(time.Millisecond)
	c0, err := config.C.NetworkServer.Band.Band.GetUplinkChannel(0)
	assert.NoError(err)

	rxInfo := gw.UplinkRXInfo{
		GatewayId: ts.Gateway.GatewayID[:],
		LoraSnr:   7,
	}
	rxInfo.Time, _ = ptypes.TimestampProto(now)
	rxInfo.TimeSinceGpsEpoch = ptypes.DurationProto(10 * time.Second)

	txInfo := gw.UplinkTXInfo{
		Frequency: uint32(c0.Frequency),
	}
	assert.NoError(helpers.SetUplinkTXInfoDataRate(&txInfo, 0, config.C.NetworkServer.Band.Band))

	testTable := []struct {
		BeforeFunc           func(ds *storage.DeviceSession)
		Name                 string
		DeviceSession        storage.DeviceSession
		PHYPayload           lorawan.PHYPayload
		ExpectedBeaconLocked bool
	}{
		{
			Name:          "trigger beacon locked",
			DeviceSession: ds,
			PHYPayload: lorawan.PHYPayload{
				MHDR: lorawan.MHDR{
					Major: lorawan.LoRaWANR1,
					MType: lorawan.UnconfirmedDataUp,
				},
				MACPayload: &lorawan.MACPayload{
					FHDR: lorawan.FHDR{
						DevAddr: ds.DevAddr,
						FCnt:    ds.FCntUp,
						FCtrl: lorawan.FCtrl{
							ClassB: true,
						},
					},
				},
			},
			ExpectedBeaconLocked: true,
		},
		{
			BeforeFunc: func(ds *storage.DeviceSession) {
				ds.BeaconLocked = true
			},
			Name:          "trigger beacon unlocked",
			DeviceSession: ds,
			PHYPayload: lorawan.PHYPayload{
				MHDR: lorawan.MHDR{
					Major: lorawan.LoRaWANR1,
					MType: lorawan.UnconfirmedDataUp,
				},
				MACPayload: &lorawan.MACPayload{
					FHDR: lorawan.FHDR{
						DevAddr: ds.DevAddr,
						FCnt:    ds.FCntUp,
						FCtrl: lorawan.FCtrl{
							ClassB: false,
						},
					},
				},
			},
			ExpectedBeaconLocked: false,
		},
	}

	for _, test := range testTable {
		ts.T().Run(test.Name, func(t *testing.T) {
			assert := require.New(t)

			if test.BeforeFunc != nil {
				test.BeforeFunc(&test.DeviceSession)
			}

			// create device-session
			assert.NoError(storage.SaveDeviceSession(ts.RedisPool(), test.DeviceSession))

			// set MIC
			assert.NoError(test.PHYPayload.SetUplinkDataMIC(lorawan.LoRaWAN1_0, 0, 0, 0, test.DeviceSession.FNwkSIntKey, test.DeviceSession.SNwkSIntKey))

			phyB, err := test.PHYPayload.MarshalBinary()
			assert.NoError(err)

			uplinkFrame := gw.UplinkFrame{
				PhyPayload: phyB,
				RxInfo:     &rxInfo,
				TxInfo:     &txInfo,
			}
			assert.NoError(uplink.HandleRXPacket(uplinkFrame))

			ds, err := storage.GetDeviceSession(ts.RedisPool(), test.DeviceSession.DevEUI)
			assert.NoError(err)

			assert.Equal(test.ExpectedBeaconLocked, ds.BeaconLocked)

			if test.ExpectedBeaconLocked {
				queueItems, err := storage.GetDeviceQueueItemsForDevEUI(config.C.PostgreSQL.DB, test.DeviceSession.DevEUI)
				assert.NoError(err)

				for _, qi := range queueItems {
					assert.NotNil(qi.EmitAtTimeSinceGPSEpoch)
					assert.NotNil(qi.TimeoutAfter)
				}
			}
		})
	}
}

func (ts *ClassBTestSuite) TestDownlink() {
	assert := require.New(ts.T())

	ts.CreateDeviceSession(storage.DeviceSession{
		DevAddr:     lorawan.DevAddr{1, 2, 3, 4},
		JoinEUI:     lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},
		FNwkSIntKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		SNwkSIntKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		NwkSEncKey:  lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		FCntUp:      8,
		NFCntDown:   5,
		UplinkGatewayHistory: map[lorawan.EUI64]storage.UplinkGatewayHistory{
			lorawan.EUI64{1, 2, 1, 2, 1, 2, 1, 2}: storage.UplinkGatewayHistory{},
		},
		EnabledUplinkChannels: []int{0, 1, 2},
		BeaconLocked:          true,
		PingSlotFrequency:     868300000,
		PingSlotDR:            2,
		RX2Frequency:          869525000,
	})

	emitTime := gps.Time(time.Now().Add(-10 * time.Second)).TimeSinceGPSEpoch()
	fPortTen := uint8(10)

	txInfo := gw.DownlinkTXInfo{
		GatewayId:         []byte{1, 2, 1, 2, 1, 2, 1, 2},
		Frequency:         uint32(ts.DeviceSession.PingSlotFrequency),
		Power:             int32(config.C.NetworkServer.Band.Band.GetDownlinkTXPower(ts.DeviceSession.PingSlotFrequency)),
		TimeSinceGpsEpoch: ptypes.DurationProto(emitTime),
	}
	assert.NoError(helpers.SetDownlinkTXInfoDataRate(&txInfo, 2, config.C.NetworkServer.Band.Band))

	txInfoDefaultFreq := txInfo
	txInfoDefaultFreq.Frequency = 869525000

	tests := []DownlinkTest{
		{
			Name:          "class-b downlink",
			DeviceSession: *ts.DeviceSession,
			DeviceQueueItems: []storage.DeviceQueueItem{
				{DevEUI: ts.Device.DevEUI, FPort: 10, FCnt: 5, FRMPayload: []byte{1, 2, 3}, EmitAtTimeSinceGPSEpoch: &emitTime},
			},
			Assert: []Assertion{
				AssertFCntUp(8),
				AssertNFCntDown(6),
				AssertDownlinkFrame(txInfo, lorawan.PHYPayload{
					MHDR: lorawan.MHDR{
						MType: lorawan.UnconfirmedDataDown,
						Major: lorawan.LoRaWANR1,
					},
					MIC: lorawan.MIC{164, 6, 172, 129},
					MACPayload: &lorawan.MACPayload{
						FHDR: lorawan.FHDR{
							DevAddr: ts.DeviceSession.DevAddr,
							FCnt:    5,
							FCtrl: lorawan.FCtrl{
								ADR: true,
							},
						},
						FPort: &fPortTen,
						FRMPayload: []lorawan.Payload{
							&lorawan.DataPayload{Bytes: []byte{1, 2, 3}},
						},
					},
				}),
			},
		},
		{
			Name:          "class-b downlink with more data",
			DeviceSession: *ts.DeviceSession,
			DeviceQueueItems: []storage.DeviceQueueItem{
				{DevEUI: ts.Device.DevEUI, FPort: 10, FCnt: 5, FRMPayload: []byte{1, 2, 3}, EmitAtTimeSinceGPSEpoch: &emitTime},
				{DevEUI: ts.Device.DevEUI, FPort: 10, FCnt: 6, FRMPayload: []byte{1, 2, 3}, EmitAtTimeSinceGPSEpoch: &emitTime},
			},
			Assert: []Assertion{
				AssertFCntUp(8),
				AssertNFCntDown(6),
				AssertDownlinkFrame(txInfo, lorawan.PHYPayload{
					MHDR: lorawan.MHDR{
						MType: lorawan.UnconfirmedDataDown,
						Major: lorawan.LoRaWANR1,
					},
					MIC: lorawan.MIC{39, 244, 225, 101},
					MACPayload: &lorawan.MACPayload{
						FHDR: lorawan.FHDR{
							DevAddr: ts.DeviceSession.DevAddr,
							FCnt:    5,
							FCtrl: lorawan.FCtrl{
								ADR:      true,
								FPending: true,
								ClassB:   true, // shares the same bit as FPending
							},
						},
						FPort: &fPortTen,
						FRMPayload: []lorawan.Payload{
							&lorawan.DataPayload{Bytes: []byte{1, 2, 3}},
						},
					},
				}),
			},
		},
		{
			BeforeFunc: func(tst *DownlinkTest) error {
				tst.DeviceSession.BeaconLocked = false
				return nil
			},
			Name:          "class-b downlink, but no beacon lock",
			DeviceSession: *ts.DeviceSession,
			DeviceQueueItems: []storage.DeviceQueueItem{
				{DevEUI: ts.Device.DevEUI, FPort: 10, FCnt: 5, FRMPayload: []byte{1, 2, 3}, EmitAtTimeSinceGPSEpoch: &emitTime},
			},
			Assert: []Assertion{
				AssertFCntUp(8),
				AssertNFCntDown(5),
			},
		},
		{
			BeforeFunc: func(tst *DownlinkTest) error {
				config.C.NetworkServer.NetworkSettings.ClassB.PingSlotFrequency = 0
				tst.DeviceSession.PingSlotFrequency = 0
				return nil
			},
			Name:          "class-b downlink, with default band frequency plan",
			DeviceSession: *ts.DeviceSession,
			DeviceQueueItems: []storage.DeviceQueueItem{
				{DevEUI: ts.Device.DevEUI, FPort: 10, FCnt: 5, FRMPayload: []byte{1, 2, 3}, EmitAtTimeSinceGPSEpoch: &emitTime},
			},
			Assert: []Assertion{
				AssertFCntUp(8),
				AssertNFCntDown(6),
				AssertDownlinkFrame(txInfoDefaultFreq, lorawan.PHYPayload{
					MHDR: lorawan.MHDR{
						MType: lorawan.UnconfirmedDataDown,
						Major: lorawan.LoRaWANR1,
					},
					MIC: lorawan.MIC{164, 6, 172, 129},
					MACPayload: &lorawan.MACPayload{
						FHDR: lorawan.FHDR{
							DevAddr: ts.DeviceSession.DevAddr,
							FCnt:    5,
							FCtrl: lorawan.FCtrl{
								ADR: true,
							},
						},
						FPort: &fPortTen,
						FRMPayload: []lorawan.Payload{
							&lorawan.DataPayload{Bytes: []byte{1, 2, 3}},
						},
					},
				}),
			},
		},
	}

	for _, tst := range tests {
		ts.T().Run(tst.Name, func(t *testing.T) {
			ts.AssertDownlinkTest(t, tst)
		})
	}
}

func TestClassB(t *testing.T) {
	suite.Run(t, new(ClassBTestSuite))
}
