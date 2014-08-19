package main

/*
 * Copyright 2014 Albert P. Tobey <atobey@datastax.com> @AlTobey
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * A minimal program for streaming sensor data from a Wii Balance Board.
 * Assumes the board is already paired with BlueZ >= 5.0,
 * xwiimote is installed, and the hid_wiimote kernel module is loaded/configured
 */

// #cgo pkg-config: libxwiimote
// #include <xwiimote.h>
import "C"

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"
	"time"
	"unsafe"

	"github.com/gocql/gocql"
)

const THRESHOLD_COUNT = 10 // number of readings in dir before recording it

var cqlFlag, ksFlag string

func init() {
	flag.StringVar(&cqlFlag, "cql", "127.0.0.1", "IP or IP:port of the Cassandra CQL service")
	flag.StringVar(&ksFlag, "ks", "f7u12", "keyspace containing the f7u12 schema")
}

func main() {
	flag.Parse()

	fmt.Printf("Connecting to Cassandra...\n")

	cluster := gocql.NewCluster(cqlFlag)
	cluster.Keyspace = ksFlag
	cluster.Consistency = gocql.One
	cass, err := cluster.CreateSession()
	if err != nil {
		panic(fmt.Sprintf("Error creating Cassandra session: %v", err))
	}
	defer cass.Close()

	fmt.Printf("Watching for Wii devices...\n")

	// the xwii monitor watches udev for new devices
	xmon := C.xwii_monitor_new(true, false)
	defer C.free(unsafe.Pointer(xmon))

	// check for new devices ever 2 seconds
	pticker := time.NewTicker(time.Second * 2)
	for _ = range pticker.C {
		// returns the sysfs path to the device as a string
		cdev := C.xwii_monitor_poll(xmon)
		gdev := C.GoString(cdev)
		C.free(unsafe.Pointer(cdev))

		// non-empty string if any device was discovered
		// might make sense to have a hash of device=>goroutine to prevent
		// forkbombing, but so far on my system it isn't a problem
		if gdev != "" {
			// kick off a goroutine for each device even though
			// there is usually just one
			go handle_device(gdev, cass)
		}
	}
}

func handle_device(dev string, cass *gocql.Session) {
	// get the Bluetooth MAC address of the device
	mac_bytes, err := ioutil.ReadFile(path.Join(dev, "..", "address"))
	if err != nil {
		log.Fatalf("Could not read MAC address: %s", err)
	}
	mac := strings.TrimSpace(string(mac_bytes))

	fmt.Printf("Found device: %s, path: %s\n", mac, dev)

	// convert to a C string to pass into xwii_iface_new
	cdev := C.CString(dev)
	defer C.free(unsafe.Pointer(cdev))

	// have xwiimote create a xwii_iface object
	var iface *C.struct_xwii_iface
	cerr := C.xwii_iface_new(&iface, cdev)
	if int(cerr) != 0 {
		log.Fatal("Failed to create xwiimote interface object for device %s\n", dev)
	}

	// open the device, only requesting balance board data
	cerr = C.xwii_iface_open(iface, C.XWII_IFACE_BALANCE_BOARD)
	if int(cerr) != 0 {
		log.Fatalf("Failed to open xwiimote interface for device %s: %d\n", dev, int(cerr))
	}

	defer C.xwii_iface_close(iface, C.XWII_IFACE_BALANCE_BOARD)
	defer C.free(unsafe.Pointer(iface))

	// create an event struct, to pass into xwii_iface_poll()
	var ev C.struct_xwii_event

	// a ring buffer of sensor data
	ring := NewBBbucket(20)

	// similar but only used for calibration
	// my board is heavy on one corner, so any time the total mass drops below
	// 300 grams, throw the values seen into this ring and grab the median to
	// subtract from new values > 300 grams
	calibration := NewBBbucket(100)
	offsets := calibration.Summarize()

	// Poll the device and process every event it sends. Place events into ring and
	// check for thresholds on every pass. Once a pair of sensors outweigh DIR_NONE
	// percent of the total mass on the board for THRESHOLD_COUNT readings, record
	// the chosen direction
	ticker := time.NewTicker(time.Millisecond)
	for now := range ticker.C {
		// poll the device
		// will try again on the next pass if it returns EAGAIN,
		// break & return on any other errors
		cerr = C.xwii_iface_poll(iface, &ev)
		if int(cerr) == -11 { // EAGAIN
			continue
		} else if int(cerr) != 0 {
			log.Printf("xwii_iface_poll failed on device %s: %d\n", dev, int(cerr))
			break
		}

		// only the balance board was requested, ignore everything else
		switch ev._type {
		case C.XWII_EVENT_BALANCE_BOARD:
			// these values are already calibrated by the hid-wiimote kernel module
			// see: https://github.com/torvalds/linux/blob/master/drivers/hid/hid-wiimote-modules.c#L1348
			vals := BBevent{}
			vals[SENSOR_RF] = int(binary.LittleEndian.Uint16(ev.v[0:4]))
			vals[SENSOR_RR] = int(binary.LittleEndian.Uint16(ev.v[12:16]))
			vals[SENSOR_LF] = int(binary.LittleEndian.Uint16(ev.v[24:28]))
			vals[SENSOR_LR] = int(binary.LittleEndian.Uint16(ev.v[36:40]))

			bbd := BBdata{now, vals, DIR_NONE}

			// don't bother counting if the total pressure is too low (nobody on the board)
			if vals.Total() < 300 {
				calibration.Insert(&bbd)
				offsets = calibration.Summarize()
				continue
			} else {
				// reduce values by the P50 of idle values
				for i, v := range bbd.Data {
					// make sure not to go negative (which wraps the value and ruins everything)
					if v > offsets.SMean[i] {
						bbd.Data[i] = v - offsets.SMean[i]
					}
				}
				ring.Insert(&bbd)
			}

			smry := ring.Summarize()

			// determine the direction chosen by finding the maximum value of each corresponding
			// sensor pair. Neutral position is preferred by dividing the sum of all 4 sensors
			// by 3 instead of taking the mean (/4)
			dirs := [5]int{}
			dirs[DIR_UP] = smry.SPercent[SENSOR_RF] + smry.SPercent[SENSOR_LF]
			dirs[DIR_DOWN] = smry.SPercent[SENSOR_RR] + smry.SPercent[SENSOR_LR]
			dirs[DIR_LEFT] = smry.SPercent[SENSOR_LF] + smry.SPercent[SENSOR_LR]
			dirs[DIR_RIGHT] = smry.SPercent[SENSOR_RF] + smry.SPercent[SENSOR_RR]
			dirs[DIR_NONE] = (dirs[DIR_UP] + dirs[DIR_DOWN] + dirs[DIR_LEFT] + dirs[DIR_RIGHT]) / 3

			for i, _ := range dirs {
				if dirs[i] >= dirs[bbd.Dir] {
					bbd.Dir = Dir(i)
				}
			}

			// make the dir count accurate since direction is guessed after computing the summary
			smry.Dirs[bbd.Dir] += 1
			// add the MAC address to the summary struct
			smry.MacAddress = mac

			if smry.Dirs[bbd.Dir] > THRESHOLD_COUNT {
				err = smry.SaveToCassandra(cass, bbd.Dir.String())
				if err != nil {
					log.Printf("Failed to write to Cassandra: %s\n", err)
				}

				fmt.Printf("% 5s: % 4d, % 4d, rf(% 4d), rr(% 4d), lf(% 4d), lr(% 4d)\n", bbd.Dir, smry.Weight, smry.Stdev, smry.SPercent[0], smry.SPercent[1], smry.SPercent[2], smry.SPercent[3])
				ring.Reset()
			}
		default:
			log.Printf("Unrecognized event type: %d\n", int(ev._type))
		}
	}
}
