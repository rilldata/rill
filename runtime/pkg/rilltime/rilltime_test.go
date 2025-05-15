package rilltime

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/pkg/timeutil"
	"github.com/stretchr/testify/require"
)

var (
	grains    = []string{"s", "m", "h", "D", "W", "M", "Q", "Y"}
	now       = "2025-05-15T10:32:36Z"
	minTime   = "2020-01-01T00:32:36Z"
	maxTime   = "2025-05-14T06:32:36Z"
	watermark = "2025-05-13T06:32:36Z"

	twoPeriodAgoStarts = []string{
		"2025-05-13T06:32:34Z", // s
		"2025-05-13T06:30:00Z", // m
		"2025-05-13T04:00:00Z", // h
		"2025-05-11T00:00:00Z", // D
		"2025-04-28T00:00:00Z", // W
		"2025-03-01T00:00:00Z", // M
		"2024-10-01T00:00:00Z", // Q
		"2023-01-01T00:00:00Z", // Y
	}
	twoPeriodAgoWeekBoundaryStarts = []string{
		"2025-05-13T06:32:34Z", // s
		"2025-05-13T06:30:00Z", // m
		"2025-05-13T04:00:00Z", // h
		"2025-05-11T00:00:00Z", // D
		"2025-04-28T00:00:00Z", // W
		"2025-03-03T00:00:00Z", // M
		"2024-09-30T00:00:00Z", // Q
		"2023-01-02T00:00:00Z", // Y
	}
	prevPeriodStarts = []string{
		"2025-05-13T06:32:35Z", // s
		"2025-05-13T06:31:00Z", // m
		"2025-05-13T05:00:00Z", // h
		"2025-05-12T00:00:00Z", // D
		"2025-05-05T00:00:00Z", // W
		"2025-04-01T00:00:00Z", // M
		"2025-01-01T00:00:00Z", // Q
		"2024-01-01T00:00:00Z", // Y
	}
	curPeriodStarts = []string{
		"2025-05-13T06:32:36Z", // s
		"2025-05-13T06:32:00Z", // m
		"2025-05-13T06:00:00Z", // h
		"2025-05-13T00:00:00Z", // D
		"2025-05-12T00:00:00Z", // W
		"2025-05-01T00:00:00Z", // M
		"2025-04-01T00:00:00Z", // Q
		"2025-01-01T00:00:00Z", // Y
	}
	curPeriodEnds = []string{
		"2025-05-13T06:32:37Z", // s
		"2025-05-13T06:33:00Z", // m
		"2025-05-13T07:00:00Z", // h
		"2025-05-14T00:00:00Z", // D
		"2025-05-19T00:00:00Z", // W
		"2025-06-01T00:00:00Z", // M
		"2025-07-01T00:00:00Z", // Q
		"2026-01-01T00:00:00Z", // Y
	}
)

func Test_CompletePreviousAndCurrentGrain(t *testing.T) {
	var testCases []testCase

	expectedGrains := []timeutil.TimeGrain{
		timeutil.TimeGrainMillisecond, // for s
		timeutil.TimeGrainSecond,      // for m
		timeutil.TimeGrainMinute,      // for h
		timeutil.TimeGrainHour,        // for D
		timeutil.TimeGrainDay,         // for W
		timeutil.TimeGrainDay,         // for M
		timeutil.TimeGrainMonth,       // for Q
		timeutil.TimeGrainMonth,       // for Y
	}

	for i, grain := range grains {
		tmGrain := timeutil.TimeGrain(i + 2)

		// Current period
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("%[1]s^ to %[1]s$", grain),
			start:     curPeriodStarts[i],
			end:       curPeriodEnds[i],
			grain:     expectedGrains[i],
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-%[1]s^ to +%[1]s$", grain),
			start:     curPeriodStarts[i],
			end:       curPeriodEnds[i],
			grain:     expectedGrains[i],
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-0%[1]s^ to +0%[1]s$", grain),
			start:     curPeriodStarts[i],
			end:       curPeriodEnds[i],
			grain:     expectedGrains[i],
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-1%[1]s$ to +1%[1]s^", grain),
			start:     curPeriodStarts[i],
			end:       curPeriodEnds[i],
			grain:     expectedGrains[i],
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-0%[1]s/%[1]s^ to +0%[1]s/%[1]s$", grain),
			start:     curPeriodStarts[i],
			end:       curPeriodEnds[i],
			grain:     expectedGrains[i],
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("1%[1]s starting %[1]s^", grain),
			start:     curPeriodStarts[i],
			end:       curPeriodEnds[i],
			grain:     expectedGrains[i],
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("1%[1]s ending %[1]s$", grain),
			start:     curPeriodStarts[i],
			end:       curPeriodEnds[i],
			grain:     expectedGrains[i],
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("%[1]s!", grain),
			start:     curPeriodStarts[i],
			end:       curPeriodEnds[i],
			grain:     expectedGrains[i],
		})

		// Previous period
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-1%[1]s^ to %[1]s^", grain),
			start:     prevPeriodStarts[i],
			end:       curPeriodStarts[i],
			grain:     expectedGrains[i],
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-1%[1]s^ to -1%[1]s$", grain),
			start:     prevPeriodStarts[i],
			end:       curPeriodStarts[i],
			grain:     expectedGrains[i],
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-2%[1]s$ to -1%[1]s$", grain),
			start:     prevPeriodStarts[i],
			end:       curPeriodStarts[i],
			grain:     expectedGrains[i],
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-1%[1]s/%[1]s^ to %[1]s/%[1]s^", grain),
			start:     prevPeriodStarts[i],
			end:       curPeriodStarts[i],
			grain:     expectedGrains[i],
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("1%[1]s starting -1%[1]s^", grain),
			start:     prevPeriodStarts[i],
			end:       curPeriodStarts[i],
			grain:     expectedGrains[i],
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("1%[1]s ending -1%[1]s$", grain),
			start:     prevPeriodStarts[i],
			end:       curPeriodStarts[i],
			grain:     expectedGrains[i],
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-1%[1]s!", grain),
			start:     prevPeriodStarts[i],
			end:       curPeriodStarts[i],
			grain:     expectedGrains[i],
		})

		// Current 3 periods
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-2%[1]s^ to %[1]s$", grain),
			start:     twoPeriodAgoStarts[i],
			end:       curPeriodEnds[i],
			grain:     tmGrain,
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("3%[1]s starting -2%[1]s^", grain),
			start:     twoPeriodAgoStarts[i],
			end:       curPeriodEnds[i],
			grain:     tmGrain,
		})

		// Previous 2 periods
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-2%[1]s^ to %[1]s^", grain),
			start:     twoPeriodAgoStarts[i],
			end:       curPeriodStarts[i],
			grain:     tmGrain,
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("2%[1]s ending %[1]s^", grain),
			start:     twoPeriodAgoStarts[i],
			end:       curPeriodStarts[i],
			grain:     tmGrain,
		})
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func Test_FirstAndLastOfPeriod(t *testing.T) {
	var testCases []testCase

	// Expected timestamps for each grain.
	// Each entry is an array starting for grain just after it, EG: for D, timestamps will start for W, M, etc.
	lastStarts := [][]string{
		{"2025-05-13T06:29:58Z", "2025-05-13T03:59:58Z"}, // for s
		{"2025-05-13T03:58:00Z", "2025-05-10T23:58:00Z"}, // for m
		{"2025-05-10T22:00:00Z", "2025-04-27T22:00:00Z"}, // for h
		{"2025-04-26T00:00:00Z", "2025-02-27T00:00:00Z"}, // for D
		{"2025-02-17T00:00:00Z", "2024-09-16T00:00:00Z"}, // for W
		{"2024-08-01T00:00:00Z", "2022-11-01T00:00:00Z"}, // for M
		{"2022-07-01T00:00:00Z"},                         // for Q
	}
	ordinalStarts := [][]string{
		{"2025-05-13T06:30:01Z", "2025-05-13T04:00:01Z"}, // for s
		{"2025-05-13T04:01:00Z", "2025-05-11T00:01:00Z"}, // for m
		{"2025-05-11T01:00:00Z", "2025-04-28T01:00:00Z"}, // for h
		{"2025-04-29T00:00:00Z", "2025-03-02T00:00:00Z"}, // for D
		{"2025-03-10T00:00:00Z", "2024-10-07T00:00:00Z"}, // for W
		{"2024-11-01T00:00:00Z", "2023-02-01T00:00:00Z"}, // for M
		{"2023-04-01T00:00:00Z"},                         // for Q
	}
	firstEnds := [][]string{
		{"2025-05-13T06:30:02Z", "2025-05-13T04:00:02Z"}, // for s
		{"2025-05-13T04:02:00Z", "2025-05-11T00:02:00Z"}, // for m
		{"2025-05-11T02:00:00Z", "2025-04-28T02:00:00Z"}, // for h
		{"2025-04-30T00:00:00Z", "2025-03-03T00:00:00Z"}, // for D
		{"2025-03-17T00:00:00Z", "2024-10-14T00:00:00Z"}, // for W
		{"2024-12-01T00:00:00Z", "2023-03-01T00:00:00Z"}, // for M
		{"2023-07-01T00:00:00Z"},                         // for Q
	}

	var grainParis []testGrainPair

	for i, grain := range grains {
		if i < len(grains)-1 {
			grainParis = append(grainParis, testGrainPair{grain, grains[i+1]})
		}
		if i < len(grains)-2 {
			grainParis = append(grainParis, testGrainPair{grain, grains[i+2]})
		}
	}
	// Select tests with higher order grains
	//grainParis = append(grainParis, testGrainPair{"D", "Y"})
	//grainParis = append(grainParis, testGrainPair{"W", "Y"})

	for _, grainPair := range grainParis {
		grainIndex := slices.Index(grains, grainPair.grain)
		higherGrainIndex := slices.Index(grains, grainPair.higherGrain)
		indexDiff := higherGrainIndex - grainIndex - 1
		tmGrain := timeutil.TimeGrain(grainIndex + 2)

		snapGrain := grainPair.higherGrain
		if grainPair.grain == "W" {
			snapGrain += "W"
		}

		lastStart := lastStarts[grainIndex][indexDiff]
		ordinalStart := ordinalStarts[grainIndex][indexDiff]
		firstEnd := firstEnds[grainIndex][indexDiff]

		higherPeriod := twoPeriodAgoStarts[higherGrainIndex]
		if grainPair.grain == "W" {
			higherPeriod = twoPeriodAgoWeekBoundaryStarts[higherGrainIndex]
		}

		// Last 2 periods of -2<higher grain>
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("2%[1]s ending -2%[2]s/%[3]s^", grainPair.grain, grainPair.higherGrain, snapGrain),
			start:     lastStart,
			end:       higherPeriod,
			grain:     tmGrain,
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-2%[2]s/%[3]s^-2%[1]s to -2%[2]s/%[3]s^", grainPair.grain, grainPair.higherGrain, snapGrain),
			start:     lastStart,
			end:       higherPeriod,
			//grain:     tmGrain, // TODO: certain cases dont correct identify grain
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-3%[2]s/%[3]s$-2%[1]s to -3%[2]s/%[3]s$", grainPair.grain, grainPair.higherGrain, snapGrain),
			start:     lastStart,
			end:       higherPeriod,
			//grain:     tmGrain, // TODO: certain cases dont correct identify grain
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf(">2%[1]s of -3%[2]s!", grainPair.grain, grainPair.higherGrain, snapGrain),
			start:     lastStart,
			end:       higherPeriod,
			grain:     tmGrain,
		})

		// First 2 periods of -2<higher grain>
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("2%[1]s starting -2%[2]s/%[3]s^", grainPair.grain, grainPair.higherGrain, snapGrain),
			start:     higherPeriod,
			end:       firstEnd,
			grain:     tmGrain,
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-2%[2]s/%[3]s^ to -2%[2]s/%[3]s^+2%[1]s", grainPair.grain, grainPair.higherGrain, snapGrain),
			start:     higherPeriod,
			end:       firstEnd,
			//grain:     tmGrain, // TODO: certain cases dont correct identify grain
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("<2%[1]s of -2%[2]s!", grainPair.grain, grainPair.higherGrain, snapGrain),
			start:     higherPeriod,
			end:       firstEnd,
			grain:     tmGrain,
		})

		// 2nd period of -2<higher grain>
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("%[1]s2 of -2%[2]s!", grainPair.grain, grainPair.higherGrain, snapGrain),
			start:     ordinalStart,
			end:       firstEnd,
			grain:     lowerOrderMap[tmGrain],
		})
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func Test_WeekCorrections(t *testing.T) {
	monthsForWeekdays := []string{
		"2024-07-01T00:00:00Z", // monday
		"2025-04-01T00:00:00Z", // tuesday
		"2025-01-01T00:00:00Z", // wednesday
		"2025-05-01T00:00:00Z", // thursday
		"2024-11-01T00:00:00Z", // friday
		"2025-03-01T00:00:00Z", // saturday
		"2024-12-01T00:00:00Z", // sunday
	}
	// W-1^, W1^, W2^ for monday as first of week
	weekBoundaries := [][]string{
		{"2024-06-24T00:00:00Z", "2024-07-01T00:00:00Z", "2024-07-08T00:00:00Z"},
		{"2025-03-24T00:00:00Z", "2025-03-31T00:00:00Z", "2025-04-07T00:00:00Z"},
		{"2024-12-23T00:00:00Z", "2024-12-30T00:00:00Z", "2025-01-06T00:00:00Z"},
		{"2025-04-21T00:00:00Z", "2025-04-28T00:00:00Z", "2025-05-05T00:00:00Z"},
		{"2024-10-28T00:00:00Z", "2024-11-04T00:00:00Z", "2024-11-11T00:00:00Z"},
		{"2025-02-24T00:00:00Z", "2025-03-03T00:00:00Z", "2025-03-10T00:00:00Z"},
		{"2024-11-25T00:00:00Z", "2024-12-02T00:00:00Z", "2024-12-09T00:00:00Z"},
	}
	// W-1^, W1^, W2^ for sunday as first of week
	weekBoundariesForSunday := [][]string{
		{"2024-06-23T00:00:00Z", "2024-06-30T00:00:00Z", "2024-07-07T00:00:00Z"},
		{"2025-03-23T00:00:00Z", "2025-03-30T00:00:00Z", "2025-04-06T00:00:00Z"},
		{"2024-12-22T00:00:00Z", "2024-12-29T00:00:00Z", "2025-01-05T00:00:00Z"},
		{"2025-04-27T00:00:00Z", "2025-05-04T00:00:00Z", "2025-05-11T00:00:00Z"},
		{"2024-10-27T00:00:00Z", "2024-11-03T00:00:00Z", "2024-11-10T00:00:00Z"},
		{"2025-02-23T00:00:00Z", "2025-03-02T00:00:00Z", "2025-03-09T00:00:00Z"},
		{"2024-11-24T00:00:00Z", "2024-12-01T00:00:00Z", "2024-12-08T00:00:00Z"},
	}

	var testCases []testCase

	for i, weekday := range monthsForWeekdays {
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("W1 as of %s", weekday),
			start:     weekBoundaries[i][1],
			end:       weekBoundaries[i][2],
			grain:     timeutil.TimeGrainDay,
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("<1W as of %s", weekday),
			start:     weekBoundaries[i][1],
			end:       weekBoundaries[i][2],
			grain:     timeutil.TimeGrainDay,
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf(">1W of -1M! as of %s", weekday),
			start:     weekBoundaries[i][0],
			end:       weekBoundaries[i][1],
			grain:     timeutil.TimeGrainDay,
		})

		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("W1 as of %s", weekday),
			start:     weekBoundariesForSunday[i][1],
			end:       weekBoundariesForSunday[i][2],
			grain:     timeutil.TimeGrainDay,
			FirstDay:  7,
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("<1W as of %s", weekday),
			start:     weekBoundariesForSunday[i][1],
			end:       weekBoundariesForSunday[i][2],
			grain:     timeutil.TimeGrainDay,
			FirstDay:  7,
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf(">1W of -1M! as of %s", weekday),
			start:     weekBoundariesForSunday[i][0],
			end:       weekBoundariesForSunday[i][1],
			grain:     timeutil.TimeGrainDay,
			FirstDay:  7,
		})
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func Test_IsoTimeRanges(t *testing.T) {
	testCases := []testCase{
		{"2025-02-20T01:23:45Z to 2025-07-15T02:34:50Z", "2025-02-20T01:23:45Z", "2025-07-15T02:34:50Z", timeutil.TimeGrainSecond, 1, 1},
		{"2025-02-20T01:23:45Z / 2025-07-15T02:34:50Z", "2025-02-20T01:23:45Z", "2025-07-15T02:34:50Z", timeutil.TimeGrainSecond, 1, 1},

		{"2025-02-20T01:23", "2025-02-20T01:23:00Z", "2025-02-20T01:24:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"2025-02-20T01", "2025-02-20T01:00:00Z", "2025-02-20T02:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"2025-02-20", "2025-02-20T00:00:00Z", "2025-02-21T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"2025-02", "2025-02-01T00:00:00Z", "2025-03-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"2025", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func Test_Eval_watermark_on_boundary(t *testing.T) {
	now := "2026-05-14T10:32:36Z"
	minTime := "2020-01-01T00:32:36Z"
	maxTime := "2025-07-01T00:00:00Z"   // month and quarter boundary
	watermark := "2025-05-12T00:00:00Z" // day and week boundary
	testCases := []testCase{
		{"-2D^ to D^", "2025-05-10T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2D$ to D$", "2025-05-11T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-4D^ to -2D^", "2025-05-08T00:00:00Z", "2025-05-10T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1D^ to D^", "2025-05-11T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"D^ to D$", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-2D^ to watermark", "2025-05-10T00:00:00Z", "2025-05-12T00:00:00.001Z", timeutil.TimeGrainDay, 1, 1},
		{"-2D^ to +1D^", "2025-05-10T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		{"-2D!", "2025-05-10T00:00:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-2W!", "2025-04-28T00:00:00Z", "2025-05-05T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2M! as of latest", "2025-05-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2Q! as of latest", "2025-01-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},

		{"H2 of -1D!", "2025-05-11T01:00:00Z", "2025-05-11T02:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"D2 of -1W!", "2025-05-06T00:00:00Z", "2025-05-07T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"W2 of -1M! as of latest", "2025-06-09T00:00:00Z", "2025-06-16T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W2 of -1Q! as of latest", "2025-04-07T00:00:00Z", "2025-04-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W2 of -1Y! as of 2024", "2023-01-09T00:00:00Z", "2023-01-16T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func Test_KatmanduTimezone(t *testing.T) {
	tz, err := time.LoadLocation("Asia/Kathmandu")
	require.NoError(t, err)

	testCases := []testCase{
		{"-2D!", "2025-05-10T18:15:00Z", "2025-05-11T18:15:00Z", timeutil.TimeGrainHour, 1, 1},
		{"D!", "2025-05-12T18:15:00Z", "2025-05-13T18:15:00Z", timeutil.TimeGrainHour, 1, 1},
		{"D^ to watermark", "2025-05-12T18:15:00Z", "2025-05-13T06:32:36.001Z", timeutil.TimeGrainHour, 1, 1},

		{"W1", "2025-04-27T18:15:00Z", "2025-05-04T18:15:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W1 of -2M!", "2025-03-02T18:15:00Z", "2025-03-09T18:15:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W1 of -1Y!", "2023-12-31T18:15:00Z", "2024-01-07T18:15:00Z", timeutil.TimeGrainDay, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, tz)
}

func Test_BackwardsCompatibility(t *testing.T) {
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

		{"inf", "2020-01-01T00:32:36Z", "2025-05-14T06:32:36.001Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"P2DT10H", "2025-05-11T20:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainHour, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func Test_EvalMisc(t *testing.T) {
	testCases := []testCase{
		// No snapping
		{"2m ending -2d", "2025-05-11T06:30:36Z", "2025-05-11T06:32:36Z", timeutil.TimeGrainMinute, 1, 1},
		{"-4d to -2d", "2025-05-09T06:32:36Z", "2025-05-11T06:32:36Z", timeutil.TimeGrainDay, 1, 1},

		// Ending on boundary explicitly
		{"Y^ to watermark", "2025-01-01T00:00:00Z", "2025-05-13T06:32:36.001Z", timeutil.TimeGrainMonth, 1, 1},
		{"Y^ to latest", "2025-01-01T00:00:00Z", "2025-05-14T06:32:36.001Z", timeutil.TimeGrainMonth, 1, 1},
		{"Y^ to now", "2025-01-01T00:00:00Z", "2025-05-15T10:32:36.001Z", timeutil.TimeGrainMonth, 1, 1},
		{"watermark to latest", "2025-05-13T06:32:36Z", "2025-05-14T06:32:36.001Z", timeutil.TimeGrainUnspecified, 1, 1},

		// `as of` without explicit truncate. Should take the higher order for calculating ordinals
		{"D2 as of -2Y", "2023-05-02T00:00:00Z", "2023-05-03T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"W2 as of -2Y", "2023-05-08T00:00:00Z", "2023-05-15T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Snapping using `/W` does not correct for ISO week boundary.
		{"-1y/W^ to -1y/W$ as of 2025-05-17T13:43:00Z", "2024-05-13T00:00:00Z", "2024-05-20T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1y/W^ to -1y/W$ as of 2025-05-15T13:43:00Z", "2024-05-13T00:00:00Z", "2024-05-20T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Snapping using `/YW` will snap by year and corrects for ISO week boundary.
		{"-2Y/YW^ to -1Y/YW^", "2023-01-02T00:00:00Z", "2024-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-2Y/YW^ to -2Y/YW$", "2023-01-02T00:00:00Z", "2024-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"Y/YW^ to W^", "2024-12-30T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},

		// The follow 2 are different. -5W4M3Q2Y applies together whereas -5W-4M-3Q-2Y applies separately.
		// This can lead to a slightly different start/end times when weeks are involved.
		{"-5W4M3Q2Y to -4W3M2Q1Y", "2022-03-09T06:32:36Z", "2023-07-16T06:32:36Z", timeutil.TimeGrainWeek, 1, 1},
		{"-5W-4M-3Q-2Y to -4W-3M-2Q-1Y", "2022-03-08T06:32:36Z", "2023-07-15T06:32:36Z", timeutil.TimeGrainMonth, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
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

func runTests(t *testing.T, testCases []testCase, now, minTime, maxTime, watermark string, tz *time.Location) {
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
				Now:        nowTm,
				MinTime:    minTimeTm,
				MaxTime:    maxTimeTm,
				Watermark:  watermarkTm,
				FirstDay:   testCase.FirstDay,
				FirstMonth: testCase.FirstMonth,
			})
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
