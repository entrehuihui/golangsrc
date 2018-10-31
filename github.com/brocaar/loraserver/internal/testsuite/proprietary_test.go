package testsuite

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/brocaar/loraserver/api/as"
	commonPB "github.com/brocaar/loraserver/api/common"
	"github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/loraserver/internal/storage"
	"github.com/brocaar/lorawan"
)

type ProprietaryTestCase struct {
	IntegrationTestSuite
}

func (ts *ProprietaryTestCase) TestDownlink() {
	tests := []DownlinkProprietaryTest{
		{
			Name: "send proprietary payload",
			SendProprietaryPayloadRequest: ns.SendProprietaryPayloadRequest{
				MacPayload:            []byte{1, 2, 3, 4},
				Mic:                   []byte{5, 6, 7, 8},
				GatewayMacs:           [][]byte{{8, 7, 6, 5, 4, 3, 2, 1}},
				PolarizationInversion: true,
				Frequency:             868100000,
				Dr:                    5,
			},

			Assert: []Assertion{
				AssertDownlinkFrame(gw.DownlinkTXInfo{
					GatewayId:   []byte{8, 7, 6, 5, 4, 3, 2, 1},
					Immediately: true,
					Frequency:   868100000,
					Power:       14,
					Modulation:  commonPB.Modulation_LORA,
					ModulationInfo: &gw.DownlinkTXInfo_LoraModulationInfo{
						LoraModulationInfo: &gw.LoRaModulationInfo{
							Bandwidth:             125,
							SpreadingFactor:       7,
							CodeRate:              "4/5",
							PolarizationInversion: true,
						},
					},
				}, lorawan.PHYPayload{
					MHDR: lorawan.MHDR{
						Major: lorawan.LoRaWANR1,
						MType: lorawan.Proprietary,
					},
					MACPayload: &lorawan.DataPayload{Bytes: []byte{1, 2, 3, 4}},
					MIC:        lorawan.MIC{5, 6, 7, 8},
				}),
			},
		},
	}

	for _, tst := range tests {
		ts.T().Run(tst.Name, func(t *testing.T) {
			ts.AssertDownlinkProprietaryTest(t, tst)
		})
	}
}

func (ts *ProprietaryTestCase) TestUplink() {
	// the routing profile is needed as the ns will send the proprietary
	// frame to all application-servers.
	ts.CreateRoutingProfile(storage.RoutingProfile{})

	ts.CreateGateway(storage.Gateway{
		GatewayID: lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Location: storage.GPSPoint{
			Latitude:  1.1234,
			Longitude: 2.345,
		},
		Altitude: 10,
	})

	tests := []UplinkProprietaryTest{
		{
			Name: "uplink proprietary payload",
			PHYPayload: lorawan.PHYPayload{
				MHDR: lorawan.MHDR{
					Major: lorawan.LoRaWANR1,
					MType: lorawan.Proprietary,
				},
				MACPayload: &lorawan.DataPayload{Bytes: []byte{1, 2, 3, 4}},
				MIC:        lorawan.MIC{5, 6, 7, 8},
			},
			TXInfo: gw.UplinkTXInfo{
				Frequency:  868100000,
				Modulation: commonPB.Modulation_LORA,
				ModulationInfo: &gw.UplinkTXInfo_LoraModulationInfo{
					LoraModulationInfo: &gw.LoRaModulationInfo{
						Bandwidth:       125,
						CodeRate:        "4/5",
						SpreadingFactor: 12,
					},
				},
			},
			RXInfo: gw.UplinkRXInfo{
				GatewayId: []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Timestamp: 12345,
				Rssi:      -10,
				LoraSnr:   5,
			},
			Assert: []Assertion{
				AssertASHandleProprietaryUplinkRequest(as.HandleProprietaryUplinkRequest{
					MacPayload: []byte{1, 2, 3, 4},
					Mic:        []byte{5, 6, 7, 8},
					TxInfo: &gw.UplinkTXInfo{
						Frequency:  868100000,
						Modulation: commonPB.Modulation_LORA,
						ModulationInfo: &gw.UplinkTXInfo_LoraModulationInfo{
							LoraModulationInfo: &gw.LoRaModulationInfo{
								Bandwidth:       125,
								SpreadingFactor: 12,
								CodeRate:        "4/5",
							},
						},
					},
					RxInfo: []*gw.UplinkRXInfo{
						{
							GatewayId: []byte{1, 2, 3, 4, 5, 6, 7, 8},
							Rssi:      -10,
							LoraSnr:   5,
							Timestamp: 12345,
							Location: &commonPB.Location{
								Latitude:  1.1234,
								Longitude: 2.345,
								Altitude:  10,
							},
						},
					},
				}),
			},
		},
	}

	for _, tst := range tests {
		ts.T().Run(tst.Name, func(t *testing.T) {
			ts.AssertUplinkProprietaryTest(t, tst)
		})
	}
}

func TestProprietary(t *testing.T) {
	suite.Run(t, new(ProprietaryTestCase))
}
