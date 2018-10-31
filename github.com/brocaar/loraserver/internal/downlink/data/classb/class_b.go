package classb

import (
	"crypto/aes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/loraserver/internal/gps"
	"github.com/brocaar/loraserver/internal/storage"
	"github.com/brocaar/lorawan"
)

const (
	beaconPeriod   = 128 * time.Second
	beaconReserved = 2120 * time.Millisecond
	beaconGuard    = 3 * time.Second
	beaconWindow   = 122880 * time.Millisecond
	pingPeriodBase = 1 << 12
	slotLen        = 30 * time.Millisecond
	scheduleMargin = 5 * time.Second
)

// GetBeaconStartForTime returns the beacon start time as a duration
// since GPS epoch for the given time.Time.
func GetBeaconStartForTime(ts time.Time) time.Duration {
	gpsTime := gps.Time(ts).TimeSinceGPSEpoch()

	return gpsTime - (gpsTime % beaconPeriod)
}

// GetPingOffset returns the ping offset for the given beacon.
func GetPingOffset(beacon time.Duration, devAddr lorawan.DevAddr, pingNb int) (int, error) {
	if pingNb == 0 {
		return 0, errors.New("pingNb must be > 0")
	}

	if beacon%beaconPeriod != 0 {
		return 0, fmt.Errorf("beacon must be a multiple of %s", beaconPeriod)
	}

	devAddrBytes, err := devAddr.MarshalBinary()
	if err != nil {
		return 0, errors.Wrap(err, "marshal devaddr error")
	}

	pingPeriod := pingPeriodBase / pingNb
	beaconTime := uint32(int64(beacon/time.Second) % (1 << 32))

	key := lorawan.AES128Key{} // 16 x 0x00
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return 0, errors.Wrap(err, "new cipher error")
	}

	if block.BlockSize() != 16 {
		return 0, errors.New("block size of 16 was expected")
	}

	b := make([]byte, len(key))
	rand := make([]byte, len(key))

	binary.LittleEndian.PutUint32(b[0:4], beaconTime)
	copy(b[4:8], devAddrBytes)
	block.Encrypt(rand, b)

	return (int(rand[0]) + int(rand[1])*256) % pingPeriod, nil
}

// GetNextPingSlotAfter returns the next pingslot occuring after the given gps epoch timestamp.
func GetNextPingSlotAfter(afterGPSEpochTS time.Duration, devAddr lorawan.DevAddr, pingNb int) (time.Duration, error) {
	if pingNb == 0 {
		return 0, errors.New("pingNb must be > 0")
	}
	beaconStart := afterGPSEpochTS - (afterGPSEpochTS % beaconPeriod)
	pingPeriod := pingPeriodBase / pingNb

	for {
		pingOffset, err := GetPingOffset(beaconStart, devAddr, pingNb)
		if err != nil {
			return 0, err
		}

		for n := 0; n < pingNb; n++ {
			gpsEpochTime := beaconStart + beaconReserved + (time.Duration(pingOffset+n*pingPeriod) * slotLen)

			if gpsEpochTime > afterGPSEpochTS {
				log.WithFields(log.Fields{
					"dev_addr":                   devAddr,
					"beacon_start_time_s":        int(beaconStart / beaconPeriod),
					"after_beacon_start_time_ms": int((gpsEpochTime - beaconStart) / time.Millisecond),
					"ping_offset_ms":             pingOffset,
					"ping_slot_n":                n,
					"ping_nb":                    pingNb,
				}).Info("get next ping-slot timestamp")
				return gpsEpochTime, nil
			}
		}

		beaconStart += beaconPeriod
	}
}

// ScheduleDeviceQueueToPingSlotsForDevEUI schedules the device-queue for the given
// DevEUI to Class-B ping slots.
func ScheduleDeviceQueueToPingSlotsForDevEUI(db sqlx.Ext, dp storage.DeviceProfile, ds storage.DeviceSession) error {
	queueItems, err := storage.GetDeviceQueueItemsForDevEUI(db, ds.DevEUI)
	if err != nil {
		return errors.Wrap(err, "get device-queue items error")
	}

	scheduleAfterGPSEpochTS := gps.Time(time.Now().Add(scheduleMargin)).TimeSinceGPSEpoch()

	for _, qi := range queueItems {
		if qi.IsPending {
			continue
		}

		gpsEpochTS, err := GetNextPingSlotAfter(scheduleAfterGPSEpochTS, ds.DevAddr, ds.PingSlotNb)
		if err != nil {
			return errors.Wrap(err, "get next ping-slot after error")
		}

		timeoutTime := time.Time(gps.NewFromTimeSinceGPSEpoch(gpsEpochTS)).Add(time.Second * time.Duration(dp.ClassBTimeout))
		qi.EmitAtTimeSinceGPSEpoch = &gpsEpochTS
		qi.TimeoutAfter = &timeoutTime

		if err := storage.UpdateDeviceQueueItem(db, &qi); err != nil {
			return errors.Wrap(err, "update device-queue item error")
		}

		scheduleAfterGPSEpochTS = gpsEpochTS
	}

	log.WithFields(log.Fields{
		"dev_eui": ds.DevEUI,
		"count":   len(queueItems),
	}).Info("device-queue items scheduled to ping-slots")

	return nil
}
