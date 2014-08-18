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
type Sensor int

const (
	DIR_UP    Dir    = 0
	DIR_DOWN  Dir    = 1
	DIR_LEFT  Dir    = 2
	DIR_RIGHT Dir    = 3
	DIR_NONE  Dir    = 4
	SENSOR_RF Sensor = 0 // right front
	SENSOR_RR Sensor = 1 // right rear
	SENSOR_LF Sensor = 2 // left front
	SENSOR_LR Sensor = 3 // left rear
)

// indexes should always use SENSOR_{RF,RR,LF,LR} constants
type BBevent [4]int

type BBdata struct {
	TS   time.Time
	Data BBevent
	Dir  Dir
}

func (bbd BBevent) Total() (total int) {
	for _, v := range bbd {
		total += v
	}
	return
}

func (bbd BBevent) Average() int {
	return bbd.Total() / 4
}

func (bbd *BBdata) String() string {
	return fmt.Sprintf("% 14d, % 4d, % 4d, % 4d, % 4d, % 6d, % 4d",
		bbd.TS.Nanosecond(), bbd.Data[SENSOR_RF], bbd.Data[SENSOR_RR], bbd.Data[SENSOR_LF], bbd.Data[SENSOR_LR],
		bbd.Data.Total(), bbd.Data.Average())
}

type BBbucket struct {
	Data  []*BBdata
	Count int
	Idx   int
	Size  int // only the requested size, don't use for computing indices!
}

type BBvals []int

func (v BBvals) Len() int           { return len(v) }
func (v BBvals) Less(i, j int) bool { return v[i] < v[j] }
func (v BBvals) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

type BBsummary struct {
	Count     int     `json:"count"`
	Period    int     `json:"period"`
	Weight    int     `json:"weight"`
	Min       int     `json:"min"`
	Max       int     `json:"max"`
	Sum       int     `json:"sum"`
	Mean      int     `json:"mean"`
	Variance  int     `json:"variance"`
	Stdev     int     `json:"stdev"`
	Dirs      [5]Dir  `json:"dirs"`
	SPercent  BBevent `json:"dist"` // distribution of weight by percent
	SSum      BBevent `json:"sums"` // individual sensors
	SMean     BBevent `json:"means"`
	SVariance BBevent `json:"variances"`
	SStdev    BBevent `json:"stdevs"`
	P5        BBevent `json:"p5"`
	P25       BBevent `json:"p25"`
	P50       BBevent `json:"p50"`
	P75       BBevent `json:"p75"`
	P95       BBevent `json:"p95"`
	First     *BBdata `json:"first"`
	Last      *BBdata `json:"last"`
}

func NewBBbucket(size int) (out BBbucket) {
	out.Size = size
	out.Data = make([]*BBdata, size)
	for i, _ := range out.Data {
		out.Data[i] = &BBdata{}
	}
	return out
}

// inserts a BBdata event into the bucket and advances the index
// by one, rolling over to 0 as needed
func (bbb *BBbucket) Insert(d *BBdata) {
	bbb.Data[bbb.Idx] = d

	// wrap around
	bbb.Idx += 1
	if bbb.Idx == len(bbb.Data) {
		bbb.Idx = 0
	}

	bbb.Count += 1
}

func (bbb *BBbucket) Reset() {
	for i, _ := range bbb.Data {
		bbb.Data[i] = &BBdata{}
	}
	bbb.Idx = 0
	bbb.Count = 0
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

	smry.First = bbb.Data[0]
	prev := bbb.Data[0]
	for i, d := range bbb.Data {
		if i > bbb.Count {
			break
		}

		total := d.Data.Total()
		smry.Sum += total
		smry.Count += 1
		smry.Last = d

		// counts by direction detected
		smry.Dirs[d.Dir] += 1

		// time elapsed since previous event
		smry.Period += (prev.TS.Nanosecond() - d.TS.Nanosecond())

		if total < smry.Min {
			smry.Min = total
		}

		if total > smry.Max {
			smry.Max = total
		}

		// populate the pivot lists and per-sensor sums
		for s, _ := range d.Data {
			pivot[s][i] = d.Data[s]
			smry.SSum[s] += d.Data[s]
		}

		prev = d
	}

	// mean time elapsed between samples
	smry.Period = smry.Period / int(smry.Count)

	smry.Mean = smry.Sum / smry.Count

	// mean for each sensor & percentage of weight on each sensor
	total := smry.SSum.Total()
	for s, v := range smry.SSum {
		smry.SMean[s] = v / smry.Count
		if total > v {
			smry.SPercent[s] = int(math.Floor((float64(v) / float64(total)) * 100))
		} else {
			fmt.Printf("total !> v: %d !> %d     (%v[%d])\n", total, v, smry.SSum, s)
		}
	}

	// distance from the mean squared for stdev
	var dsum int
	var dsums BBevent
	for _, d := range bbb.Data {
		diff := d.Data.Total() - smry.Mean
		dsum += diff * diff

		for s, _ := range d.Data {
			sdiff := d.Data[s] - smry.SMean[s]
			dsums[s] += sdiff * sdiff
		}
	}

	// variance & stdev
	smry.Variance = dsum / smry.Count
	smry.Stdev = int(math.Ceil(math.Sqrt(float64(smry.Variance))))
	for s, _ := range dsums {
		smry.SVariance[s] = dsums[s] / smry.Count
		smry.SStdev[s] = int(math.Ceil(math.Sqrt(float64(smry.SVariance[s]))))
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

	// fill in weight using the P50 value
	smry.Weight = int(math.Floor(float64(smry.P50[0]+smry.P50[1]+smry.P50[2]+smry.P50[3]) * 0.022046226218))

	return smry
}

func (d Dir) String() string {
	if d == DIR_UP {
		return "up"
	} else if d == DIR_DOWN {
		return "down"
	} else if d == DIR_LEFT {
		return "left"
	} else if d == DIR_RIGHT {
		return "right"
	} else {
		return "none"
	}
}

func (s Sensor) String() string {
	switch s {
	case SENSOR_RF:
		return "right-front"
	case SENSOR_RR:
		return "right-rear"
	case SENSOR_LF:
		return "left-front"
	case SENSOR_LR:
		return "left-rear"
	}
	return "ERROR"
}
