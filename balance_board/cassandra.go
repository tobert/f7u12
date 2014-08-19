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
 */

import (
	"time"

	"github.com/gocql/gocql"
)

func (smry *BBsummary) SaveToCassandra(cass *gocql.Session, direction string) error {
	query := `
INSERT INTO balance_board_input
(bucket, mac, ts, direction,
 count, period, weight, min, max, sum, mean, variance, stdev,
 rf_pcnt, rr_pcnt, lf_pcnt, lr_pcnt,
 rf_mean, rr_mean, lf_mean, lr_mean,
 rf_stdev, rr_stdev, lf_stdev, lr_stdev
) VALUES (?,?,?,?, ?,?,?,?,?,?,?,?,?, ?,?,?,?, ?,?,?,?, ?,?,?,?)
`

	bucket := smry.Timestamp.Truncate(time.Minute)

	return cass.Query(query,
		bucket, smry.MacAddress, smry.Timestamp, direction,
		smry.Count, smry.Period, smry.Weight, smry.Min, smry.Max, smry.Sum, smry.Mean, smry.Variance, smry.Stdev,
		smry.SPercent[SENSOR_RF], smry.SPercent[SENSOR_RR], smry.SPercent[SENSOR_LF], smry.SPercent[SENSOR_LR],
		smry.SMean[SENSOR_RF], smry.SMean[SENSOR_RR], smry.SMean[SENSOR_LF], smry.SMean[SENSOR_LR],
		smry.SStdev[SENSOR_RF], smry.SStdev[SENSOR_RR], smry.SStdev[SENSOR_LF], smry.SStdev[SENSOR_LR],
	).Exec()
}
