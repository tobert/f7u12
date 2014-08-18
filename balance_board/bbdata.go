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
 * Statistical functions for analyzing balance board data streams.
 */

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type Dir int

func (d Dir) String() string {
	if d == DIR_UP {
		return "up"
	} else if (d == DIR_DOWN) {
		return "down"
	} else if (d == DIR_LEFT) {
		return "left"
	} else if (d == DIR_RIGHT) {
		return "right"
	} else {
		return "none"
	}
}

const (
	DIR_UP    Dir = 0
	DIR_DOWN  Dir = 1
	DIR_LEFT  Dir = 2
	DIR_RIGHT Dir = 3
	DIR_NONE  Dir = 4
)

type BBsensor [4]uint32

type BBdata struct {
	TS time.Time // timestamp
	// right front / right rear / left front / left rear
	Sensor BBsensor
	Dir    Dir
}

func (bbd BBsensor) Total() (total uint32) {
	for _, v := range bbd {
		total += v
	}
	return
}

func (bbd BBsensor) Average() uint32 {
	return bbd.Total() / 4
}

func (bbd *BBdata) String() string {
	return fmt.Sprintf("% 14d, % 4d, % 4d, % 4d, % 4d, % 6d, % 4d",
		bbd.TS.Nanosecond(), bbd.Sensor[0], bbd.Sensor[1], bbd.Sensor[2], bbd.Sensor[3],
		bbd.Sensor.Total(), bbd.Sensor.Average())
}

type BBbucket struct {
	Data []BBdata
	Idx  int
	Size int // only the requested size, don't use for computing indices!
}

type BBvals []uint32

func (v BBvals) Len() int           { return len(v) }
func (v BBvals) Less(i, j int) bool { return v[i] < v[j] }
func (v BBvals) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

type BBsummary struct {
	Count     uint32   `json:"count"`
	Period    int      `json:"period"`
	Min       uint32   `json:"min"`
	Max       uint32   `json:"max"`
	Sum       uint32   `json:"sum"`
	Mean      uint32   `json:"mean"`
	Variance  uint32   `json:"variance"`
	Stdev     uint32   `json:"stdev"`
	Dist      BBsensor `json:"dist"` // distribution of weight by percent
	Dirs      [5]Dir   `json:"dirs"`
	SSum      BBsensor `json:"sums"` // individual sensors
	SMean     BBsensor `json:"means"`
	SVariance BBsensor `json:"variances"`
	SStdev    BBsensor `json:"stdevs"`
	P5        BBsensor `json:"p5"`
	P25       BBsensor `json:"p25"`
	P50       BBsensor `json:"p50"`
	P75       BBsensor `json:"p75"`
	P95       BBsensor `json:"p95"`
	First     BBdata   `json:"first"`
	Last      BBdata   `json:"last"`
}

func NewBBbucket(size int) (out BBbucket) {
	out.Size = size
	out.Data = make([]BBdata, size)
	return out
}

// inserts a BBdata event into the bucket and advances the index
// by one, rolling over to 0 as needed
func (bbb *BBbucket) Insert(d BBdata) int {
	// wrap around
	if bbb.Idx == len(bbb.Data)-1 {
		bbb.Idx = 0
	} else {
		bbb.Idx += 1
	}

	bbb.Data[bbb.Idx] = d

	return bbb.Idx
}

// window into a bucket that is another bucket
func (bbb *BBbucket) Window(start, end int) (out BBbucket) {
	out.Data = bbb.Data[start:end]
	return out
}

// summarize the data in the bucket
func (bbb *BBbucket) Summarize() (smry BBsummary) {
	// populated with the raw values to be sorted to get percentiles
	pivot := []BBvals{
		make(BBvals, len(bbb.Data)), // right front
		make(BBvals, len(bbb.Data)), // right rear
		make(BBvals, len(bbb.Data)), // left front
		make(BBvals, len(bbb.Data)), // left rear
	}

	smry.Count = uint32(len(bbb.Data))
	smry.First = bbb.Data[0]
	smry.Last = bbb.Data[len(bbb.Data)-1]

	prev := &bbb.Data[0]
	for i, d := range bbb.Data {
		total := d.Sensor.Total()
		smry.Sum += total

		// time elapsed since previous event
		smry.Period += (prev.TS.Nanosecond() - d.TS.Nanosecond())

		if total < smry.Min {
			smry.Min = total
		}

		if total > smry.Max {
			smry.Max = total
		}

		// populate the pivot lists and per-sensor sums
		for s, _ := range d.Sensor {
			pivot[s][i] = d.Sensor[s]
			smry.SSum[s] += d.Sensor[s]
		}

		smry.Dirs[d.Dir] += 1

		prev = &d
	}

	// mean time elapsed between samples
	smry.Period = smry.Period / int(smry.Count)
	smry.Mean = smry.Sum / smry.Count

	// mean for each sensor & percentage of weight on each sensor
	for s, _ := range smry.SSum {
		smry.SMean[s] = smry.SSum[s] / smry.Count
		smry.Dist[s] = uint32(math.Ceil((float64(smry.SSum[s]) / float64(smry.SSum.Total())) * 100))
	}

	// distance from the mean squared for stdev
	var dsum uint32
	var dsums BBsensor
	for _, d := range bbb.Data {
		diff := d.Sensor.Total() - smry.Mean
		dsum += diff * diff

		for s, _ := range d.Sensor {
			sdiff := d.Sensor[s] - smry.SMean[s]
			dsums[s] += sdiff * sdiff
		}
	}

	// variance & stdev
	smry.Variance = dsum / smry.Count
	smry.Stdev = uint32(math.Ceil(math.Sqrt(float64(smry.Variance))))
	for s, _ := range dsums {
		smry.SVariance[s] = dsums[s] / smry.Count
		smry.SStdev[s] = uint32(math.Ceil(math.Sqrt(float64(smry.SVariance[s]))))
	}

	// percentiles
	for i, sensor := range pivot {
		sort.Sort(sensor)

		idx := func(pc float64) int {
			idx := math.Floor(float64(len(sensor)) * (pc / 100))
			return int(idx)
		}

		smry.P5[i] = sensor[idx(5)]
		smry.P25[i] = sensor[idx(25)]
		smry.P50[i] = sensor[idx(50)]
		smry.P75[i] = sensor[idx(75)]
		smry.P95[i] = sensor[idx(95)]
	}

	return smry
}
