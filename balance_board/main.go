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
	"fmt"
	"log"
	"time"
	"unsafe"
)

func main() {
	fmt.Printf("Started\n")

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
			go handle_device(gdev)
		}
	}
}

func handle_device(dev string) {
	fmt.Printf("Found device: %s\n", dev)
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

	// might make sense to put a timer on this
	for {
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
			// convert the pressure sensor data to integers
			a := binary.LittleEndian.Uint32(ev.v[0:4])   // right front
			b := binary.LittleEndian.Uint32(ev.v[12:16]) // right rear
			c := binary.LittleEndian.Uint32(ev.v[24:28]) // left front
			d := binary.LittleEndian.Uint32(ev.v[36:40]) // left rear

			// TODO: pass to a function that converts to direction
			fmt.Printf("% 4d, % 4d, % 4d, % 4d, % 6d\n", a, b, c, d, (a + b + c + d))
		default:
			log.Printf("Unrecognized event type: %d\n", int(ev._type))
		}
	}
}
