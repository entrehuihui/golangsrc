package mqtthandler

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/pkg/errors"
)

type alarmInfo struct {
	humidity_low     bool
	humidity_high    bool
	temperature_low  bool
	temperature_high bool
	electricity_low  bool
}

type humiture struct {
	temp    float64
	hum     float64
	elec    float64
	up_date time.Time
	alarm   byte
}

func unmarshalAlarm(alarm byte) alarmInfo {
	var info alarmInfo
	if alarm&0x01 == 0x01 {
		info.humidity_high = true
	}
	if alarm&0x02 == 0x02 {
		info.temperature_high = true
	}
	if alarm&0x04 == 0x04 {
		info.humidity_low = true
	}
	if alarm&0x08 == 0x08 {
		info.temperature_low = true
	}
	if alarm&0x10 == 0x10 {
		info.electricity_low = true
	}
	return info
}

func (a alarmInfo) String() string {
	buf := bytes.NewBufferString("")
	if a.humidity_high {
		buf.WriteString("湿度过高;")
	}
	if a.humidity_low {
		buf.WriteString("湿度过低;")
	}
	if a.temperature_high {
		buf.WriteString("温度过高;")
	}
	if a.temperature_low {
		buf.WriteString("温度过低;")
	}
	if a.electricity_low {
		buf.WriteString("电量过低;")
	}
	return string(buf.Bytes())
}

// 解码温湿度，版本日期2018-10-23
//  frame_header function_code  timestamp  humidity     temperature  electricity  alarm   frame_tail
//      1             1             4         1          2                1         1          1
func decodeHumiture(data []byte) (humiture, error) {
	var (
		start  int
		end    int
		result = humiture{}
	)
	if len(data) < 12 {
		return result, errors.New("humiture data format error.")
	}
	if data[1] != 0x01 {
		return result, errors.New("not single humiture data.")
	}

	start = 2
	end = start + 4
	_up_date := binary.BigEndian.Uint32(data[start:end])
	result.up_date = time.Unix(int64(_up_date), 0)

	start = end
	end += 1
	result.hum = float64(data[start])

	start = end
	end += 2
	_temp := binary.BigEndian.Uint16(data[start:end])
	temp := float64(int16(_temp)) / 10.0
	result.temp = temp

	start = end
	end += 1
	result.elec = float64(data[start])

	start = end
	end = end + 1
	result.alarm = data[start]

	return result, nil
}

/*
// 批量上行数据解释，请见相关协议文档
帧头 |功能码 | 分包数 | 实际数据包数量 ，
分包序号1，长度，时间，
分包序号2，长度，温度数据，
分包序号3，长度，湿度数据，
分包序号4，长度 ，电量数据，
分包序号5，长度 ，报警数据
*/
func decodeHumitures(data []byte) ([]humiture, error) {
	if data[1] != 0x02 {
		return nil, errors.New("not multiple humiture data.")
	}
	var (
		start       int
		end         int
		packetCount int                         // 分包数
		dataCount   int                         // 实际数据包长度
		packetNum   byte                        //分包序号
		dataLen     int                         // 单包数据长度
		duration    time.Duration = time.Minute //时间间隔，固定1分钟
		startTemp   int16
		startHum    byte
		startElec   byte
		startAlarm  byte
	)
	// 分包数
	start = 2
	end = start + 1
	packetCount = int(data[start])

	// 实际数据包数量
	start = end
	end += 1
	dataCount = int(data[start])
	if dataCount == 0 {
		return nil, errors.New("实际数据包长度为0")
	}

	resultTime := make([]time.Time, 0, dataCount)
	resultTemp := make([]float64, 0, dataCount)
	resultHum := make([]float64, 0, dataCount)
	resultElec := make([]float64, 0, dataCount)
	resultAlarm := make([]byte, 0, dataCount)
	results := make([]humiture, 0, dataCount)

	// 第一个分包序号
	start = end
	end += 1
	packetNum = data[start]

	for count := 0; count < packetCount; count++ {
		switch packetNum {
		// 时间数据
		case 1:
			// 取数据长度
			start = end
			end += 1
			dataLen = int(data[start])
			if dataLen+end > len(data) {
				return nil, errors.New("时间分包数据格式错误.")
			}
			// 4字节时间数据
			start = end
			end += 4
			_upTime := binary.BigEndian.Uint32(data[start:end])
			startTime := time.Unix(int64(_upTime), 0)
			for i := 0; i < dataCount; i++ {
				resultTime = append(resultTime, startTime)
				startTime = startTime.Add(duration)
			}
			//取下一个分包序号
			start = end
			end += 1
			if start < len(data) {
				packetNum = data[start]
			}

		// 温度数据
		case 2:
			// 取数据长度
			start = end
			end += 1
			dataLen = int(data[start])
			if dataLen+end > len(data) {
				return nil, errors.New("温度分包数据格式错误.")
			}
			// 开始处理分包数据
			start = end
			for i := 0; i < dataLen; {
				// 字节A开始,为取上一次数据，详见协议文档
				if 0xa0&data[start] == 0xa0 {
					end += 1
					i += 1
					num := 0x0f & data[start]
					for j := 0; j < int(num); j++ {
						resultTemp = append(resultTemp, float64(startTemp)/10.0)
					}
					start = end
				} else {
					end += 2
					i += 2
					startTemp = int16(binary.BigEndian.Uint16(data[start:end]))
					resultTemp = append(resultTemp, float64(startTemp)/10.0)
					start = end
				}
			}

			//取下一个分包序号
			start = end
			end = start + 1
			if start < len(data) {
				packetNum = data[start]
			}

		// 湿度数据
		case 3:
			// 取数据长度
			start = end
			end += 1
			dataLen = int(data[start])
			if dataLen+end > len(data) {
				return nil, errors.New("湿度分包数据格式错误.")
			}
			// 开始处理分包数据
			start = end
			for i := 0; i < dataLen; i++ {
				// 字节A开始,为取上一次数据，详见协议文档
				if 0xa0&data[start] == 0xa0 {
					end += 1
					num := 0x0f & data[start]
					for j := 0; j < int(num); j++ {
						resultHum = append(resultHum, float64(startHum))
					}
					start = end
				} else {
					end += 1
					startHum = data[start]
					resultHum = append(resultHum, float64(startHum))
					start = end
				}
			}

			//取下一个分包序号
			start = end
			end = start + 1
			if start < len(data) {
				packetNum = data[start]
			}

		// 电量数据
		case 4:
			// 取数据长度
			start = end
			end += 1
			dataLen = int(data[start])
			if dataLen+end > len(data) {
				return nil, errors.New("电量分包数据格式错误.")
			}
			// 开始处理分包数据
			start = end
			for i := 0; i < dataLen; i++ {
				// 字节A开始,为取上一次数据，详见协议文档
				if 0xa0&data[start] == 0xa0 {
					end += 1
					num := 0x0f & data[start]
					for j := 0; j < int(num); j++ {
						resultElec = append(resultElec, float64(startElec))
					}
					start = end
				} else {
					end += 1
					startElec = data[start]
					resultElec = append(resultElec, float64(startElec))
					start = end
				}
			}

			//取下一个分包序号
			start = end
			end = start + 1
			if start < len(data) {
				packetNum = data[start]
			}

		// 报警数据
		case 5:
			// 取数据长度
			start = end
			end += 1
			dataLen = int(data[start])
			if dataLen+end > len(data) {
				return nil, errors.New("报警分包数据格式错误.")
			}
			// 开始处理分包数据
			start = end
			for i := 0; i < dataLen; i++ {
				if 0xa0&data[start] == 0xa0 {
					end += 1
					num := 0x0f & data[start]
					for j := 0; j < int(num); j++ {
						resultAlarm = append(resultAlarm, startAlarm)
					}
					start = end
				} else {
					end += 1
					startAlarm = data[start]
					resultAlarm = append(resultAlarm, startAlarm)
					start = end
				}
			}

			//取下一个分包序号
			start = end
			end = start + 1
			if start < len(data) {
				packetNum = data[start]
			}

		}
	}

	if dataCount > len(resultTemp) {
		dataCount = len(resultTemp)
	}
	if dataCount > len(resultHum) {
		dataCount = len(resultHum)
	}
	if dataCount > len(resultElec) {
		dataCount = len(resultElec)
	}
	if dataCount > len(resultTime) {
		dataCount = len(resultTime)
	}
	if dataCount > len(resultAlarm) {
		dataCount = len(resultAlarm)
	}
	for i := 0; i < dataCount; i++ {
		result := humiture{
			temp:    resultTemp[i],
			hum:     resultHum[i],
			elec:    resultElec[i],
			up_date: resultTime[i],
			alarm:   resultAlarm[i],
		}
		results = append(results, result)
	}
	return results, nil
}
