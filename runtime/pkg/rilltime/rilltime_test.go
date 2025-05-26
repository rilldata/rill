package rilltime

import (
	"fmt"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/pkg/timeutil"
	"github.com/stretchr/testify/require"
)

var (
	now       = "2025-05-15T10:32:36Z"
	minTime   = "2020-01-01T00:32:36Z"
	maxTime   = "2025-05-14T06:32:36Z"
	watermark = "2025-05-13T06:32:36Z"
)

func TestEval_PreviousAndCurrentCompleteGrain(t *testing.T) {
	testCases := []testCase{
		// Previous complete second
		{"1s in s!", "2025-05-13T06:32:35Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"-1s^ to s^", "2025-05-13T06:32:35Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"-1s^ to -1s$", "2025-05-13T06:32:35Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"-2s$ to -1s$", "2025-05-13T06:32:35Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"-1s/s^ to s/s^", "2025-05-13T06:32:35Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"1s starting -1s^", "2025-05-13T06:32:35Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"1s ending -1s$", "2025-05-13T06:32:35Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"-1s!", "2025-05-13T06:32:35Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainMillisecond, 1, 1},
		// Last 2 seconds, including current second
		{"2s in s", "2025-05-13T06:32:35Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainSecond, 1, 1},
		{"-1s^ to s$", "2025-05-13T06:32:35Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainSecond, 1, 1},
		{"2s starting -1s^", "2025-05-13T06:32:35Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainSecond, 1, 1},
		// Last 2 seconds, excluding current second
		{"2s in s!", "2025-05-13T06:32:34Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainSecond, 1, 1},
		{"-2s^ to s^", "2025-05-13T06:32:34Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainSecond, 1, 1},
		{"2s ending s^", "2025-05-13T06:32:34Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainSecond, 1, 1},
		// Current complete second
		{"1s in s", "2025-05-13T06:32:36Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"s^ to s$", "2025-05-13T06:32:36Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"-s^ to +s$", "2025-05-13T06:32:36Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"-0s^ to +0s$", "2025-05-13T06:32:36Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"-1s$ to +1s^", "2025-05-13T06:32:36Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"-0s/s^ to +0s/s$", "2025-05-13T06:32:36Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"1s starting s^", "2025-05-13T06:32:36Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"1s ending s$", "2025-05-13T06:32:36Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"s!", "2025-05-13T06:32:36Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainMillisecond, 1, 1},

		// Previous complete minute
		{"1m in m!", "2025-05-13T06:31:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"-1m^ to m^", "2025-05-13T06:31:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"-1m^ to -1m$", "2025-05-13T06:31:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"-2m$ to -1m$", "2025-05-13T06:31:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"-1m/m^ to m/m^", "2025-05-13T06:31:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"1m starting -1m^", "2025-05-13T06:31:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"1m ending -1m$", "2025-05-13T06:31:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"-1m!", "2025-05-13T06:31:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainSecond, 1, 1},
		// Last 2 minutes, including current minute
		{"2m in m", "2025-05-13T06:31:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-1m^ to m$", "2025-05-13T06:31:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"2m starting -1m^", "2025-05-13T06:31:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainMinute, 1, 1},
		// Last 2 minutes, excluding current minute
		{"2m in m!", "2025-05-13T06:30:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-2m^ to m^", "2025-05-13T06:30:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"2m ending m^", "2025-05-13T06:30:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainMinute, 1, 1},
		// Current complete minute
		{"1m in m", "2025-05-13T06:32:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"m^ to m$", "2025-05-13T06:32:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"-m^ to +m$", "2025-05-13T06:32:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"-0m^ to +0m$", "2025-05-13T06:32:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"-1m$ to +1m^", "2025-05-13T06:32:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"-0m/m^ to +0m/m$", "2025-05-13T06:32:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"1m starting m^", "2025-05-13T06:32:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"1m ending m$", "2025-05-13T06:32:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"m!", "2025-05-13T06:32:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainSecond, 1, 1},

		// Previous complete hour
		{"1h in h!", "2025-05-13T05:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-1h^ to h^", "2025-05-13T05:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-1h^ to -1h$", "2025-05-13T05:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-2h$ to -1h$", "2025-05-13T05:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-1h/h^ to h/h^", "2025-05-13T05:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"1h starting -1h^", "2025-05-13T05:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"1h ending -1h$", "2025-05-13T05:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-1h!", "2025-05-13T05:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		// Last 2 hours, including current hour
		{"2h in h", "2025-05-13T05:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-1h^ to h$", "2025-05-13T05:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"2h starting -1h^", "2025-05-13T05:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainHour, 1, 1},
		// Last 2 hours, excluding current hour
		{"2h in h!", "2025-05-13T04:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-2h^ to h^", "2025-05-13T04:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"2h ending h^", "2025-05-13T04:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainHour, 1, 1},
		// Current complete hour
		{"1h in h", "2025-05-13T06:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"h^ to h$", "2025-05-13T06:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-h^ to +h$", "2025-05-13T06:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-0h^ to +0h$", "2025-05-13T06:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-1h$ to +1h^", "2025-05-13T06:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-0h/h^ to +0h/h$", "2025-05-13T06:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"1h starting h^", "2025-05-13T06:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"1h ending h$", "2025-05-13T06:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"h!", "2025-05-13T06:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainMinute, 1, 1},

		// Previous complete day
		{"1D in D!", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-1D^ to D^", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-1D^ to -1D$", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-2D$ to -1D$", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-1D/D^ to D/D^", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"1D starting -1D^", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"1D ending -1D$", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-1D!", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		// Last 2 days, including current day
		{"2D in D", "2025-05-12T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1D^ to D$", "2025-05-12T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"2D starting -1D^", "2025-05-12T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Last 2 days, excluding current day
		{"2D in D!", "2025-05-11T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2D^ to D^", "2025-05-11T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"2D ending D^", "2025-05-11T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Current complete day
		{"1D in D", "2025-05-13T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"D^ to D$", "2025-05-13T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-D^ to +D$", "2025-05-13T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-0D^ to +0D$", "2025-05-13T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-1D$ to +1D^", "2025-05-13T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-0D/D^ to +0D/D$", "2025-05-13T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"1D starting D^", "2025-05-13T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"1D ending D$", "2025-05-13T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"D!", "2025-05-13T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainHour, 1, 1},

		// Previous complete week
		{"-1W^ to W^", "2025-05-05T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1W^ to -1W$", "2025-05-05T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2W$ to -1W$", "2025-05-05T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1W/W^ to W/W^", "2025-05-05T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"1W starting -1W^", "2025-05-05T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"1W ending -1W$", "2025-05-05T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1W!", "2025-05-05T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Last 2 weeks, including current week
		{"-1W^ to W$", "2025-05-05T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		{"2W starting -1W^", "2025-05-05T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		// Last 2 weeks, excluding current week
		{"-2W^ to W^", "2025-04-28T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		{"2W ending W^", "2025-04-28T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		// Current complete week
		{"W^ to W$", "2025-05-12T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-W^ to +W$", "2025-05-12T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-0W^ to +0W$", "2025-05-12T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1W$ to +1W^", "2025-05-12T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-0W/W^ to +0W/W$", "2025-05-12T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"1W starting W^", "2025-05-12T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"1W ending W$", "2025-05-12T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W!", "2025-05-12T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Previous complete month
		{"-1M^ to M^", "2025-04-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1M^ to -1M$", "2025-04-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2M$ to -1M$", "2025-04-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1M/M^ to M/M^", "2025-04-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"1M starting -1M^", "2025-04-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"1M ending -1M$", "2025-04-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1M!", "2025-04-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Last 2 months, including current month
		{"-1M^ to M$", "2025-04-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"2M starting -1M^", "2025-04-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		// Last 2 months, excluding current month
		{"-2M^ to M^", "2025-03-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"2M ending M^", "2025-03-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		// Current complete month
		{"M^ to M$", "2025-05-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-M^ to +M$", "2025-05-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-0M^ to +0M$", "2025-05-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1M$ to +1M^", "2025-05-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-0M/M^ to +0M/M$", "2025-05-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"1M starting M^", "2025-05-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"1M ending M$", "2025-05-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"M!", "2025-05-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Previous complete quarter
		{"-1Q^ to Q^", "2025-01-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-1Q^ to -1Q$", "2025-01-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-2Q$ to -1Q$", "2025-01-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-1Q/Q^ to Q/Q^", "2025-01-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"1Q starting -1Q^", "2025-01-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"1Q ending -1Q$", "2025-01-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-1Q!", "2025-01-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		// Last 2 quarters, including current quarter
		{"-1Q^ to Q$", "2025-01-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainQuarter, 1, 1},
		{"2Q starting -1Q^", "2025-01-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainQuarter, 1, 1},
		// Last 2 quarters, excluding current quarter
		{"-2Q^ to Q^", "2024-10-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainQuarter, 1, 1},
		{"2Q ending Q^", "2024-10-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainQuarter, 1, 1},
		// Current complete quarter
		{"Q^ to Q$", "2025-04-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-Q^ to +Q$", "2025-04-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-0Q^ to +0Q$", "2025-04-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-1Q$ to +1Q^", "2025-04-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-0Q/Q^ to +0Q/Q$", "2025-04-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"1Q starting Q^", "2025-04-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"1Q ending Q$", "2025-04-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"Q!", "2025-04-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},

		// Previous complete year
		{"-1Y^ to Y^", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-1Y^ to -1Y$", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-2Y$ to -1Y$", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-1Y/Y^ to Y/Y^", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"1Y starting -1Y^", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"1Y ending -1Y$", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-1Y!", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		// Last 2 years, including current year
		{"-1Y^ to Y$", "2024-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainYear, 1, 1},
		{"2Y starting -1Y^", "2024-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainYear, 1, 1},
		// Last 2 years, excluding current year
		{"-2Y^ to Y^", "2023-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainYear, 1, 1},
		{"2Y ending Y^", "2023-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainYear, 1, 1},
		// Current complete year
		{"Y^ to Y$", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-Y^ to +Y$", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-0Y^ to +0Y$", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-1Y$ to +1Y^", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-0Y/Y^ to +0Y/Y$", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"1Y starting Y^", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"1Y ending Y$", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"Y!", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil, timeutil.TimeGrainUnspecified)
}

func TestEval_FirstAndLastOfPeriod(t *testing.T) {
	testCases := []testCase{
		// Last 2 secs of last 4 mins
		{"2s ending -2m/m^", "2025-05-13T06:29:58Z", "2025-05-13T06:30:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"-2m/m^-2s to -2m/m^", "2025-05-13T06:29:58Z", "2025-05-13T06:30:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"-3m/m$-2s to -3m/m$", "2025-05-13T06:29:58Z", "2025-05-13T06:30:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{">2s of -3m!", "2025-05-13T06:29:58Z", "2025-05-13T06:30:00Z", timeutil.TimeGrainSecond, 1, 1},
		// First 2 secs of last 2 mins
		{"2s starting -2m/m^", "2025-05-13T06:30:00Z", "2025-05-13T06:30:02Z", timeutil.TimeGrainSecond, 1, 1},
		{"-2m/m^ to -2m/m^+2s", "2025-05-13T06:30:00Z", "2025-05-13T06:30:02Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"<2s of -2m!", "2025-05-13T06:30:00Z", "2025-05-13T06:30:02Z", timeutil.TimeGrainSecond, 1, 1},
		// Sec 2 of last 2 mins
		{"s2 of -2m!", "2025-05-13T06:30:01Z", "2025-05-13T06:30:02Z", timeutil.TimeGrainMillisecond, 1, 1},

		// Last 2 secs of last 4 hrs
		{"2s ending -2h/h^", "2025-05-13T03:59:58Z", "2025-05-13T04:00:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"-2h/h^-2s to -2h/h^", "2025-05-13T03:59:58Z", "2025-05-13T04:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"-3h/h$-2s to -3h/h$", "2025-05-13T03:59:58Z", "2025-05-13T04:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{">2s of -3h!", "2025-05-13T03:59:58Z", "2025-05-13T04:00:00Z", timeutil.TimeGrainSecond, 1, 1},
		// First 2 secs of last 2 hrs
		{"2s starting -2h/h^", "2025-05-13T04:00:00Z", "2025-05-13T04:00:02Z", timeutil.TimeGrainSecond, 1, 1},
		{"-2h/h^ to -2h/h^+2s", "2025-05-13T04:00:00Z", "2025-05-13T04:00:02Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"<2s of -2h!", "2025-05-13T04:00:00Z", "2025-05-13T04:00:02Z", timeutil.TimeGrainSecond, 1, 1},
		// Sec 2 of last 2 hrs
		{"s2 of -2h!", "2025-05-13T04:00:01Z", "2025-05-13T04:00:02Z", timeutil.TimeGrainMillisecond, 1, 1},

		// Last 2 mins of last 4 hrs
		{"2m ending -2h/h^", "2025-05-13T03:58:00Z", "2025-05-13T04:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-2h/h^-2m to -2h/h^", "2025-05-13T03:58:00Z", "2025-05-13T04:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"-3h/h$-2m to -3h/h$", "2025-05-13T03:58:00Z", "2025-05-13T04:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{">2m of -3h!", "2025-05-13T03:58:00Z", "2025-05-13T04:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		// First 2 mins of last 2 hrs
		{"2m starting -2h/h^", "2025-05-13T04:00:00Z", "2025-05-13T04:02:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-2h/h^ to -2h/h^+2m", "2025-05-13T04:00:00Z", "2025-05-13T04:02:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"<2m of -2h!", "2025-05-13T04:00:00Z", "2025-05-13T04:02:00Z", timeutil.TimeGrainMinute, 1, 1},
		// Min 2 of last 2 hrs
		{"m2 of -2h!", "2025-05-13T04:01:00Z", "2025-05-13T04:02:00Z", timeutil.TimeGrainSecond, 1, 1},

		// Last 2 mins of last 4 days
		{"2m ending -2D/D^", "2025-05-10T23:58:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-2D/D^-2m to -2D/D^", "2025-05-10T23:58:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"-3D/D$-2m to -3D/D$", "2025-05-10T23:58:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{">2m of -3D!", "2025-05-10T23:58:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		// First 2 mins of last 2 days
		{"2m starting -2D/D^", "2025-05-11T00:00:00Z", "2025-05-11T00:02:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-2D/D^ to -2D/D^+2m", "2025-05-11T00:00:00Z", "2025-05-11T00:02:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"<2m of -2D!", "2025-05-11T00:00:00Z", "2025-05-11T00:02:00Z", timeutil.TimeGrainMinute, 1, 1},
		// Min 2 of last 2 days
		{"m2 of -2D!", "2025-05-11T00:01:00Z", "2025-05-11T00:02:00Z", timeutil.TimeGrainSecond, 1, 1},

		// Last 2 hrs of last 4 days
		{"2h ending -2D/D^", "2025-05-10T22:00:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-2D/D^-2h to -2D/D^", "2025-05-10T22:00:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"-3D/D$-2h to -3D/D$", "2025-05-10T22:00:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{">2h of -3D!", "2025-05-10T22:00:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		// First 2 hrs of last 2 days
		{"2h starting -2D/D^", "2025-05-11T00:00:00Z", "2025-05-11T02:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-2D/D^ to -2D/D^+2h", "2025-05-11T00:00:00Z", "2025-05-11T02:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"<2h of -2D!", "2025-05-11T00:00:00Z", "2025-05-11T02:00:00Z", timeutil.TimeGrainHour, 1, 1},
		// Hour 2 of last 2 days
		{"h2 of -2D!", "2025-05-11T01:00:00Z", "2025-05-11T02:00:00Z", timeutil.TimeGrainMinute, 1, 1},

		// Last 2 hrs of last 4 weeks
		{"2h ending -2W/W^", "2025-04-27T22:00:00Z", "2025-04-28T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-2W/W^-2h to -2W/W^", "2025-04-27T22:00:00Z", "2025-04-28T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"-3W/W$-2h to -3W/W$", "2025-04-27T22:00:00Z", "2025-04-28T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{">2h of -3W!", "2025-04-27T22:00:00Z", "2025-04-28T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		// First 2 hrs of last 2 weeks
		{"2h starting -2W/W^", "2025-04-28T00:00:00Z", "2025-04-28T02:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-2W/W^ to -2W/W^+2h", "2025-04-28T00:00:00Z", "2025-04-28T02:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"<2h of -2W!", "2025-04-28T00:00:00Z", "2025-04-28T02:00:00Z", timeutil.TimeGrainHour, 1, 1},
		// Hour 2 of last 2 weeks
		{"h2 of -2W!", "2025-04-28T01:00:00Z", "2025-04-28T02:00:00Z", timeutil.TimeGrainMinute, 1, 1},

		// Last 2 days of last 4 weeks
		{"2D ending -2W/W^", "2025-04-26T00:00:00Z", "2025-04-28T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2W/W^-2D to -2W/W^", "2025-04-26T00:00:00Z", "2025-04-28T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"-3W/W$-2D to -3W/W$", "2025-04-26T00:00:00Z", "2025-04-28T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{">2D of -3W!", "2025-04-26T00:00:00Z", "2025-04-28T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// First 2 days of last 2 weeks
		{"2D starting -2W/W^", "2025-04-28T00:00:00Z", "2025-04-30T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2W/W^ to -2W/W^+2D", "2025-04-28T00:00:00Z", "2025-04-30T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"<2D of -2W!", "2025-04-28T00:00:00Z", "2025-04-30T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Day 2 of last 2 weeks
		{"D2 of -2W!", "2025-04-29T00:00:00Z", "2025-04-30T00:00:00Z", timeutil.TimeGrainHour, 1, 1},

		// Last 2 days of last 4 months
		{"2D ending -2M/M^", "2025-02-27T00:00:00Z", "2025-03-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2M/M^-2D to -2M/M^", "2025-02-27T00:00:00Z", "2025-03-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"-3M/M$-2D to -3M/M$", "2025-02-27T00:00:00Z", "2025-03-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{">2D of -3M!", "2025-02-27T00:00:00Z", "2025-03-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// First 2 days of last 2 months
		{"2D starting -2M/M^", "2025-03-01T00:00:00Z", "2025-03-03T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2M/M^ to -2M/M^+2D", "2025-03-01T00:00:00Z", "2025-03-03T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"<2D of -2M!", "2025-03-01T00:00:00Z", "2025-03-03T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Day 2 of last 2 months
		{"D2 of -2M!", "2025-03-02T00:00:00Z", "2025-03-03T00:00:00Z", timeutil.TimeGrainHour, 1, 1},

		// Last 2 weeks of last 4 months
		{"2W ending -2M/M/W^", "2025-02-17T00:00:00Z", "2025-03-03T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		{"-2M/M/W^-2W to -2M/M/W^", "2025-02-17T00:00:00Z", "2025-03-03T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"-3M/M/W$-2W to -3M/M/W$", "2025-02-17T00:00:00Z", "2025-03-03T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{">2W of -3M!", "2025-02-17T00:00:00Z", "2025-03-03T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		// First 2 weeks of last 2 months
		{"2W starting -2M/M/W^", "2025-03-03T00:00:00Z", "2025-03-17T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		{"-2M/M/W^ to -2M/M/W^+2W", "2025-03-03T00:00:00Z", "2025-03-17T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"<2W of -2M!", "2025-03-03T00:00:00Z", "2025-03-17T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		// Week 2 of last 2 months
		{"W2 of -2M!", "2025-03-10T00:00:00Z", "2025-03-17T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Last 2 weeks of last 4 quarters
		{"2W ending -2Q/Q/W^", "2024-09-16T00:00:00Z", "2024-09-30T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		{"-2Q/Q/W^-2W to -2Q/Q/W^", "2024-09-16T00:00:00Z", "2024-09-30T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"-3Q/Q/W$-2W to -3Q/Q/W$", "2024-09-16T00:00:00Z", "2024-09-30T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{">2W of -3Q!", "2024-09-16T00:00:00Z", "2024-09-30T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		// First 2 weeks of last 2 quarters
		{"2W starting -2Q/Q/W^", "2024-09-30T00:00:00Z", "2024-10-14T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		{"-2Q/Q/W^ to -2Q/Q/W^+2W", "2024-09-30T00:00:00Z", "2024-10-14T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"<2W of -2Q!", "2024-09-30T00:00:00Z", "2024-10-14T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		// Week 2 of last 2 quarters
		{"W2 of -2Q!", "2024-10-07T00:00:00Z", "2024-10-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Last 2 months of last 4 quarters
		{"2M ending -2Q/Q^", "2024-08-01T00:00:00Z", "2024-10-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-2Q/Q^-2M to -2Q/Q^", "2024-08-01T00:00:00Z", "2024-10-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"-3Q/Q$-2M to -3Q/Q$", "2024-08-01T00:00:00Z", "2024-10-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{">2M of -3Q!", "2024-08-01T00:00:00Z", "2024-10-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		// First 2 months of last 2 quarters
		{"2M starting -2Q/Q^", "2024-10-01T00:00:00Z", "2024-12-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-2Q/Q^ to -2Q/Q^+2M", "2024-10-01T00:00:00Z", "2024-12-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"<2M of -2Q!", "2024-10-01T00:00:00Z", "2024-12-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		// Month 2 of last 2 quarters
		{"M2 of -2Q!", "2024-11-01T00:00:00Z", "2024-12-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Last 2 months of last 4 years
		{"2M ending -2Y/Y^", "2022-11-01T00:00:00Z", "2023-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-2Y/Y^-2M to -2Y/Y^", "2022-11-01T00:00:00Z", "2023-01-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"-3Y/Y$-2M to -3Y/Y$", "2022-11-01T00:00:00Z", "2023-01-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{">2M of -3Y!", "2022-11-01T00:00:00Z", "2023-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		// First 2 months of last 2 years
		{"2M starting -2Y/Y^", "2023-01-01T00:00:00Z", "2023-03-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-2Y/Y^ to -2Y/Y^+2M", "2023-01-01T00:00:00Z", "2023-03-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"<2M of -2Y!", "2023-01-01T00:00:00Z", "2023-03-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		// Month 2 of last 2 years
		{"M2 of -2Y!", "2023-02-01T00:00:00Z", "2023-03-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Last 2 quarters of last 4 years
		{"2Q ending -2Y/Y^", "2022-07-01T00:00:00Z", "2023-01-01T00:00:00Z", timeutil.TimeGrainQuarter, 1, 1},
		{"-2Y/Y^-2Q to -2Y/Y^", "2022-07-01T00:00:00Z", "2023-01-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"-3Y/Y$-2Q to -3Y/Y$", "2022-07-01T00:00:00Z", "2023-01-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{">2Q of -3Y!", "2022-07-01T00:00:00Z", "2023-01-01T00:00:00Z", timeutil.TimeGrainQuarter, 1, 1},
		// First 2 quarters of last 2 years
		{"2Q starting -2Y/Y^", "2023-01-01T00:00:00Z", "2023-07-01T00:00:00Z", timeutil.TimeGrainQuarter, 1, 1},
		{"-2Y/Y^ to -2Y/Y^+2Q", "2023-01-01T00:00:00Z", "2023-07-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"<2Q of -2Y!", "2023-01-01T00:00:00Z", "2023-07-01T00:00:00Z", timeutil.TimeGrainQuarter, 1, 1},
		// Quarter 2 of last 2 years
		{"Q2 of -2Y!", "2023-04-01T00:00:00Z", "2023-07-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil, timeutil.TimeGrainUnspecified)
}

func TestEval_WeekCorrections(t *testing.T) {
	testCases := []testCase{
		// Boundary on Monday, week starts on Monday
		{"W1 as of 2024-07-01T00:00:00Z", "2024-07-01T00:00:00Z", "2024-07-08T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"<1W as of 2024-07-01T00:00:00Z", "2024-07-01T00:00:00Z", "2024-07-08T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{">1W of -1M! as of 2024-07-01T00:00:00Z", "2024-06-24T00:00:00Z", "2024-07-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Boundary on Monday, week starts on Sunday
		{"W1 as of 2024-07-01T00:00:00Z", "2024-06-30T00:00:00Z", "2024-07-07T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
		{"<1W as of 2024-07-01T00:00:00Z", "2024-06-30T00:00:00Z", "2024-07-07T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
		{">1W of -1M! as of 2024-07-01T00:00:00Z", "2024-06-23T00:00:00Z", "2024-06-30T00:00:00Z", timeutil.TimeGrainDay, 7, 1},

		// Boundary on Tuesday, week starts on Monday
		{"W1 as of 2025-04-01T00:00:00Z", "2025-03-31T00:00:00Z", "2025-04-07T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"<1W as of 2025-04-01T00:00:00Z", "2025-03-31T00:00:00Z", "2025-04-07T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{">1W of -1M! as of 2025-04-01T00:00:00Z", "2025-03-24T00:00:00Z", "2025-03-31T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Boundary on Tuesday, week starts on Sunday
		{"W1 as of 2025-04-01T00:00:00Z", "2025-03-30T00:00:00Z", "2025-04-06T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
		{"<1W as of 2025-04-01T00:00:00Z", "2025-03-30T00:00:00Z", "2025-04-06T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
		{">1W of -1M! as of 2025-04-01T00:00:00Z", "2025-03-23T00:00:00Z", "2025-03-30T00:00:00Z", timeutil.TimeGrainDay, 7, 1},

		// Boundary on Wednesday, week starts on Monday
		{"W1 as of 2025-01-01T00:00:00Z", "2024-12-30T00:00:00Z", "2025-01-06T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"<1W as of 2025-01-01T00:00:00Z", "2024-12-30T00:00:00Z", "2025-01-06T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{">1W of -1M! as of 2025-01-01T00:00:00Z", "2024-12-23T00:00:00Z", "2024-12-30T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Boundary on Wednesday, week starts on Sunday
		{"W1 as of 2025-01-01T00:00:00Z", "2024-12-29T00:00:00Z", "2025-01-05T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
		{"<1W as of 2025-01-01T00:00:00Z", "2024-12-29T00:00:00Z", "2025-01-05T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
		{">1W of -1M! as of 2025-01-01T00:00:00Z", "2024-12-22T00:00:00Z", "2024-12-29T00:00:00Z", timeutil.TimeGrainDay, 7, 1},

		// Boundary on Thursday, week starts on Monday
		{"W1 as of 2025-05-01T00:00:00Z", "2025-04-28T00:00:00Z", "2025-05-05T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"<1W as of 2025-05-01T00:00:00Z", "2025-04-28T00:00:00Z", "2025-05-05T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{">1W of -1M! as of 2025-05-01T00:00:00Z", "2025-04-21T00:00:00Z", "2025-04-28T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Boundary on Thursday, week starts on Sunday
		{"W1 as of 2025-05-01T00:00:00Z", "2025-05-04T00:00:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
		{"<1W as of 2025-05-01T00:00:00Z", "2025-05-04T00:00:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
		{">1W of -1M! as of 2025-05-01T00:00:00Z", "2025-04-27T00:00:00Z", "2025-05-04T00:00:00Z", timeutil.TimeGrainDay, 7, 1},

		// Boundary on Friday, week starts on Monday
		{"W1 as of 2024-11-01T00:00:00Z", "2024-11-04T00:00:00Z", "2024-11-11T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"<1W as of 2024-11-01T00:00:00Z", "2024-11-04T00:00:00Z", "2024-11-11T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{">1W of -1M! as of 2024-11-01T00:00:00Z", "2024-10-28T00:00:00Z", "2024-11-04T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Boundary on Friday, week starts on Sunday
		{"W1 as of 2024-11-01T00:00:00Z", "2024-11-03T00:00:00Z", "2024-11-10T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
		{"<1W as of 2024-11-01T00:00:00Z", "2024-11-03T00:00:00Z", "2024-11-10T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
		{">1W of -1M! as of 2024-11-01T00:00:00Z", "2024-10-27T00:00:00Z", "2024-11-03T00:00:00Z", timeutil.TimeGrainDay, 7, 1},

		// Boundary on Saturday, week starts on Monday
		{"W1 as of 2025-03-01T00:00:00Z", "2025-03-03T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"<1W as of 2025-03-01T00:00:00Z", "2025-03-03T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{">1W of -1M! as of 2025-03-01T00:00:00Z", "2025-02-24T00:00:00Z", "2025-03-03T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Boundary on Saturday, week starts on Sunday
		{"W1 as of 2025-03-01T00:00:00Z", "2025-03-02T00:00:00Z", "2025-03-09T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
		{"<1W as of 2025-03-01T00:00:00Z", "2025-03-02T00:00:00Z", "2025-03-09T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
		{">1W of -1M! as of 2025-03-01T00:00:00Z", "2025-02-23T00:00:00Z", "2025-03-02T00:00:00Z", timeutil.TimeGrainDay, 7, 1},

		// Boundary on Sunday, week starts on Monday
		{"W1 as of 2024-12-01T00:00:00Z", "2024-12-02T00:00:00Z", "2024-12-09T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"<1W as of 2024-12-01T00:00:00Z", "2024-12-02T00:00:00Z", "2024-12-09T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{">1W of -1M! as of 2024-12-01T00:00:00Z", "2024-11-25T00:00:00Z", "2024-12-02T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Boundary on Sunday, week starts on Sunday
		{"W1 as of 2024-12-01T00:00:00Z", "2024-12-01T00:00:00Z", "2024-12-08T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
		{"<1W as of 2024-12-01T00:00:00Z", "2024-12-01T00:00:00Z", "2024-12-08T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
		{">1W of -1M! as of 2024-12-01T00:00:00Z", "2024-11-24T00:00:00Z", "2024-12-01T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil, timeutil.TimeGrainUnspecified)
}

func TestEval_ShorthandSyntax(t *testing.T) {
	testCases := []testCase{
		{"7D", "2025-05-06T07:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"7D!", "2025-05-06T06:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"7D in D", "2025-05-07T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"7D in D!", "2025-05-06T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		{"MTD", "2025-05-01T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"MTD!", "2025-05-01T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"MTD in h", "2025-05-01T00:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"MTD in h!", "2025-05-01T00:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"MTD as of -1Y", "2024-05-01T00:00:00Z", "2024-05-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil, timeutil.TimeGrainUnspecified)
}

func TestEval_IsoTimeRanges(t *testing.T) {
	testCases := []testCase{
		{"2025-02-20T01:23:45Z to 2025-07-15T02:34:50Z", "2025-02-20T01:23:45Z", "2025-07-15T02:34:50Z", timeutil.TimeGrainSecond, 1, 1},
		{"2025-02-20T01:23:45Z / 2025-07-15T02:34:50Z", "2025-02-20T01:23:45Z", "2025-07-15T02:34:50Z", timeutil.TimeGrainSecond, 1, 1},

		{"2025-02-20T01:23", "2025-02-20T01:23:00Z", "2025-02-20T01:24:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"2025-02-20T01", "2025-02-20T01:00:00Z", "2025-02-20T02:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"2025-02-20", "2025-02-20T00:00:00Z", "2025-02-21T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"2025-02", "2025-02-01T00:00:00Z", "2025-03-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"2025", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil, timeutil.TimeGrainUnspecified)
}

func TestEval_WatermarkOnBoundary(t *testing.T) {
	maxTimeOnBoundary := "2025-07-01T00:00:00Z"   // month and quarter boundary
	watermarkOnBoundary := "2025-05-12T00:00:00Z" // day and week boundary
	testCases := []testCase{
		{"h^ to h$", "2025-05-13T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"D^ to h$", "2025-05-13T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainMillisecond, 1, 1},

		{"-2D^ to D^", "2025-05-11T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Simulates comparison for the above
		{"-2D^ to D^ as of -2D", "2025-05-09T00:00:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2D$ to D$", "2025-05-11T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Simulates comparison for the above
		{"-2D$ to D$ as of -2D", "2025-05-09T00:00:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"2D", "2025-05-11T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"2D in D", "2025-05-11T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-4D^ to -2D^", "2025-05-09T00:00:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1D^ to D^", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"D^ to D$", "2025-05-13T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"-2D^ to watermark", "2025-05-11T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-2D^ to +1D^", "2025-05-11T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		{"-2D!", "2025-05-11T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-2W!", "2025-04-28T00:00:00Z", "2025-05-05T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2M! as of latest", "2025-05-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2Q! as of latest", "2025-01-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},

		{"H2 of -1D!", "2025-05-12T01:00:00Z", "2025-05-12T02:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"D2 of -1W!", "2025-05-06T00:00:00Z", "2025-05-07T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"W2 of -1M! as of latest", "2025-06-09T00:00:00Z", "2025-06-16T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W2 of -1Q! as of latest", "2025-04-07T00:00:00Z", "2025-04-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W2 of -1Y! as of 2024", "2023-01-09T00:00:00Z", "2023-01-16T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTimeOnBoundary, watermarkOnBoundary, nil, timeutil.TimeGrainDay)
}

func Test_KatmanduTimezone(t *testing.T) {
	tz, err := time.LoadLocation("Asia/Kathmandu")
	require.NoError(t, err)

	testCases := []testCase{
		{"-2D!", "2025-05-10T18:15:00Z", "2025-05-11T18:15:00Z", timeutil.TimeGrainHour, 1, 1},
		{"D!", "2025-05-12T18:15:00Z", "2025-05-13T18:15:00Z", timeutil.TimeGrainHour, 1, 1},
		{"D^ to watermark", "2025-05-12T18:15:00Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainHour, 1, 1},

		{"W1", "2025-04-27T18:15:00Z", "2025-05-04T18:15:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W1 of -2M!", "2025-03-02T18:15:00Z", "2025-03-09T18:15:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W1 of -1Y!", "2023-12-31T18:15:00Z", "2024-01-07T18:15:00Z", timeutil.TimeGrainDay, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, tz, timeutil.TimeGrainUnspecified)
}

func TestEval_BackwardsCompatibility(t *testing.T) {
	testCases := []testCase{
		{"rill-TD", "2025-05-13T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"rill-WTD", "2025-05-12T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"rill-MTD", "2025-05-01T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"rill-QTD", "2025-04-01T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"rill-YTD", "2025-01-01T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},

		{"rill-PDC", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"rill-PWC", "2025-05-05T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"rill-PMC", "2025-04-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"rill-PQC", "2025-01-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"rill-PYC", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},

		// `inf` => `earliest to now`, where `now` is adjusted ref.
		{"inf", "2020-01-01T00:32:36Z", "2025-05-13T06:32:36.001Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"P2DT10H", "2025-05-11T20:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainHour, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil, timeutil.TimeGrainUnspecified)
}

func TestEval_Misc(t *testing.T) {
	testCases := []testCase{
		// Ending on boundary explicitly
		{"Y^ to watermark", "2025-01-01T00:00:00Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainMonth, 1, 1},
		{"Y^ to latest", "2025-01-01T00:00:00Z", "2025-05-14T06:32:36Z", timeutil.TimeGrainMonth, 1, 1},
		// Now is adjusted ref. Since min_grain is unspecified it defaults to millisecond
		{"Y^ to now", "2025-01-01T00:00:00Z", "2025-05-13T06:32:36.001Z", timeutil.TimeGrainMonth, 1, 1},
		{"watermark to latest", "2025-05-13T06:32:36Z", "2025-05-14T06:32:36Z", timeutil.TimeGrainUnspecified, 1, 1},

		// `as of` without explicit truncate. Should take the higher order for calculating ordinals
		{"D2 as of -2Y", "2023-05-02T00:00:00Z", "2023-05-03T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"W2 as of -2Y", "2023-05-08T00:00:00Z", "2023-05-15T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Snapping using `/W` does not correct for ISO week boundary.
		{"-1y/W^ to -1y/W$ as of 2025-05-17T13:43:00Z", "2024-05-13T00:00:00Z", "2024-05-20T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1y/W^ to -1y/W$ as of 2025-05-15T13:43:00Z", "2024-05-13T00:00:00Z", "2024-05-20T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Snapping using `/Y/W` will snap by year and corrects for ISO week boundary.
		{"-2Y/Y/W^ to -1Y/Y/W^", "2023-01-02T00:00:00Z", "2024-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-2Y/Y/W^ to -2Y/Y/W$", "2023-01-02T00:00:00Z", "2024-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"Y/Y/W^ to W^", "2024-12-30T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},

		// The following 2 ranges are different. -5W4M3Q2Y applies together whereas -5W-4M-3Q-2Y applies separately.
		// This can lead to a slightly different start/end times when weeks are involved.
		{"-5W4M3Q2Y to -4W3M2Q1Y", "2022-03-09T06:32:36.001Z", "2023-07-16T06:32:36.001Z", timeutil.TimeGrainWeek, 1, 1},
		{"-5W-4M-3Q-2Y to -4W-3M-2Q-1Y", "2022-03-08T06:32:36.001Z", "2023-07-15T06:32:36.001Z", timeutil.TimeGrainMonth, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil, timeutil.TimeGrainUnspecified)
}

func TestEval_SyntaxErrors(t *testing.T) {
	testCases := []struct {
		timeRange string
		errorMsg  string
	}{
		{"D2 of -2Y", `unexpected token "<EOF>" (expected <to> PointInTime)`},
		{"-4d", `unexpected token "<EOF>" (expected <interval>)`},
		{"D", `unexpected token "<EOF>" (expected <interval>)`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.timeRange, func(t *testing.T) {
			_, err := Parse(testCase.timeRange, ParseOptions{})
			require.Error(t, err)
			require.ErrorContains(t, err, testCase.errorMsg)
		})
	}
}

type testCase struct {
	timeRange  string
	start      string
	end        string
	grain      timeutil.TimeGrain
	FirstDay   int
	FirstMonth int
}

type testGrainPair struct {
	grain, higherGrain string
}

func runTests(t *testing.T, testCases []testCase, now, minTime, maxTime, watermark string, tz *time.Location, minTg timeutil.TimeGrain) {
	nowTm := parseTestTime(t, now)
	minTimeTm := parseTestTime(t, minTime)
	maxTimeTm := parseTestTime(t, maxTime)
	watermarkTm := parseTestTime(t, watermark)

	for _, testCase := range testCases {
		t.Run(testCase.timeRange, func(t *testing.T) {
			rt, err := Parse(testCase.timeRange, ParseOptions{
				TimeZoneOverride: tz,
			})
			require.NoError(t, err)

			start, end, grain := rt.Eval(EvalOptions{
				Now:           nowTm,
				MinTime:       minTimeTm,
				MaxTime:       maxTimeTm,
				Watermark:     watermarkTm,
				FirstDay:      testCase.FirstDay,
				FirstMonth:    testCase.FirstMonth,
				SmallestGrain: minTg,
			})
			fmt.Println(start, end)
			require.Equal(t, parseTestTime(t, testCase.start), start)
			require.Equal(t, parseTestTime(t, testCase.end), end)
			if testCase.grain != timeutil.TimeGrainUnspecified {
				require.Equal(t, testCase.grain, grain)
			}
		})
	}
}

func parseTestTime(tst *testing.T, t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return ts
}
