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

		// Previous period
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-1%[1]s^ to %[1]s^", grain),
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

		// Previous 3 periods
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

	lastStarts := []string{
		"2025-05-13T06:29:58Z", // for s
		"2025-05-13T03:58:00Z", // for m
		"2025-05-10T22:00:00Z", // for h
		"2025-04-26T00:00:00Z", // for D
		"2025-02-17T00:00:00Z", // for W
		"2024-08-01T00:00:00Z", // for M
		"2022-07-01T00:00:00Z", // for Q
	}
	ordinalStarts := []string{
		"2025-05-13T06:30:01Z", // for s
		"2025-05-13T04:01:00Z", // for m
		"2025-05-11T01:00:00Z", // for h
		"2025-04-29T00:00:00Z", // for D
		"2025-03-10T00:00:00Z", // for W
		"2024-11-01T00:00:00Z", // for M
		"2023-04-01T00:00:00Z", // for Q
	}
	firstEnds := []string{
		"2025-05-13T06:30:02Z", // for s
		"2025-05-13T04:02:00Z", // for m
		"2025-05-11T02:00:00Z", // for h
		"2025-04-30T00:00:00Z", // for D
		"2025-03-17T00:00:00Z", // for W
		"2024-12-01T00:00:00Z", // for M
		"2023-07-01T00:00:00Z", // for Q
	}

	var grainParis []testGrainPair

	for i, grain := range grains {
		if i == len(grains)-1 {
			continue
		}
		higherPeriod := twoPeriodAgoStarts[i+1]
		if grain == "W" {
			higherPeriod = "2025-03-03T00:00:00Z"
		}
		grainParis = append(grainParis, testGrainPair{grain, grains[i+1], lastStarts[i], higherPeriod, ordinalStarts[i], firstEnds[i]})
	}
	grainParis = append(grainParis, testGrainPair{"D", "M", "2025-02-27T00:00:00Z", "2025-03-01T00:00:00Z", "2025-03-02T00:00:00Z", "2025-03-03T00:00:00Z"})
	grainParis = append(grainParis, testGrainPair{"D", "Y", "2022-12-30T00:00:00Z", "2023-01-01T00:00:00Z", "2023-01-02T00:00:00Z", "2023-01-03T00:00:00Z"})
	grainParis = append(grainParis, testGrainPair{"W", "Q", "2024-09-16T00:00:00Z", "2024-09-30T00:00:00Z", "2024-10-07T00:00:00Z", "2024-10-14T00:00:00Z"})
	grainParis = append(grainParis, testGrainPair{"W", "Y", "2022-12-19T00:00:00Z", "2023-01-02T00:00:00Z", "2023-01-09T00:00:00Z", "2023-01-16T00:00:00Z"})

	for _, grainPair := range grainParis {
		grainIndex := slices.Index(grains, grainPair.grain)
		tmGrain := timeutil.TimeGrain(grainIndex + 2)

		snapGrain := grainPair.higherGrain
		if grainPair.grain == "W" {
			snapGrain += "W"
		}

		// Last 2 periods of -2<higher grain>
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("2%[1]s ending -2%[2]s^", grainPair.grain, grainPair.higherGrain),
			start:     grainPair.lastStart,
			end:       grainPair.higherPeriod,
			grain:     tmGrain,
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-2%[2]s/%[3]s^-2%[1]s to -2%[2]s/%[3]s^", grainPair.grain, grainPair.higherGrain, snapGrain),
			start:     grainPair.lastStart,
			end:       grainPair.higherPeriod,
			grain:     tmGrain,
		})

		// First 2 periods of -2<higher grain>
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("2%[1]s starting -2%[2]s^", grainPair.grain, grainPair.higherGrain),
			start:     grainPair.higherPeriod,
			end:       grainPair.firstEnd,
			grain:     tmGrain,
		})
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("-2%[2]s/%[3]s^ to -2%[2]s/%[3]s^+2%[1]s", grainPair.grain, grainPair.higherGrain, snapGrain),
			start:     grainPair.higherPeriod,
			end:       grainPair.firstEnd,
			grain:     tmGrain,
		})

		// 2nd period of -2<higher grain>
		testCases = append(testCases, testCase{
			timeRange: fmt.Sprintf("%[1]s2 of -2%[2]s!", grainPair.grain, grainPair.higherGrain, snapGrain),
			start:     grainPair.ordinalStart,
			end:       grainPair.firstEnd,
			grain:     lowerOrderMap[tmGrain],
		})
	}

	// Misc tests that do not follow exact patterns
	testCases = append(testCases, testCase{"D2 as of -2Y", "2023-05-02T00:00:00Z", "2023-05-03T00:00:00Z", timeutil.TimeGrainHour})
	testCases = append(testCases, testCase{"W2 as of -2Y", "2023-05-08T00:00:00Z", "2023-05-15T00:00:00Z", timeutil.TimeGrainDay})

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func Test_IsoTimeRanges(t *testing.T) {
	testCases := []testCase{
		{"2025-02-20T01:23:45Z to 2025-07-15T02:34:50Z", "2025-02-20T01:23:45Z", "2025-07-15T02:34:50Z", timeutil.TimeGrainSecond},
		{"2025-02-20T01:23:45Z / 2025-07-15T02:34:50Z", "2025-02-20T01:23:45Z", "2025-07-15T02:34:50Z", timeutil.TimeGrainSecond},

		{"2025-02-20T01:23", "2025-02-20T01:23:00Z", "2025-02-20T01:24:00Z", timeutil.TimeGrainSecond},
		{"2025-02-20T01", "2025-02-20T01:00:00Z", "2025-02-20T02:00:00Z", timeutil.TimeGrainMinute},
		{"2025-02-20", "2025-02-20T00:00:00Z", "2025-02-21T00:00:00Z", timeutil.TimeGrainHour},
		{"2025-02", "2025-02-01T00:00:00Z", "2025-03-01T00:00:00Z", timeutil.TimeGrainDay},
		{"2025", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func Test_EvalFinal(t *testing.T) {
	now := parseTestTime(t, "2025-03-12T10:32:36Z")
	minTime := parseTestTime(t, "2020-01-01T00:32:36Z")
	maxTime := parseTestTime(t, "2025-03-11T06:32:36Z")
	watermark := parseTestTime(t, "2025-05-13T06:32:36Z")
	testCases := []struct {
		timeRange string
		start     string
		end       string
		grain     timeutil.TimeGrain
	}{
		{"D2 as of -2Y", "2023-05-02T00:00:00Z", "2023-05-03T00:00:00Z", timeutil.TimeGrainHour},
		{"W2 as of -2Y", "2023-05-08T00:00:00Z", "2023-05-15T00:00:00Z", timeutil.TimeGrainDay},

		{"<6h of -1D!", "2025-03-09T00:00:00Z", "2025-03-09T06:00:00Z", timeutil.TimeGrainHour},
		{"-1d^ to -1d^+6h", "2025-03-09T00:00:00Z", "2025-03-09T06:00:00Z", timeutil.TimeGrainHour},
		{"6h starting -1d^", "2025-03-09T00:00:00Z", "2025-03-09T06:00:00Z", timeutil.TimeGrainHour},
		{"-1d/d^ to -1d/d^+6h", "2025-03-09T00:00:00Z", "2025-03-09T06:00:00Z", timeutil.TimeGrainHour},

		{"M^ to d^", "2025-03-01T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay},
		{"-0M/M^ to -0d/d^", "2025-03-01T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay},

		{"-4W^ to -3W^", "2025-02-10T00:00:00Z", "2025-02-17T00:00:00Z", timeutil.TimeGrainDay},
		{"-4W!", "2025-02-10T00:00:00Z", "2025-02-17T00:00:00Z", timeutil.TimeGrainDay},
		{"1W starting -4W^", "2025-02-10T00:00:00Z", "2025-02-17T00:00:00Z", timeutil.TimeGrainDay},
		{"1W ending -3W^", "2025-02-10T00:00:00Z", "2025-02-17T00:00:00Z", timeutil.TimeGrainDay},
		{"-4w/w^ to -3w/w^", "2025-02-10T00:00:00Z", "2025-02-17T00:00:00Z", timeutil.TimeGrainDay},

		{"-4Y^ to -1M^", "2021-01-01T00:00:00Z", "2025-02-01T00:00:00Z", timeutil.TimeGrainMonth},

		{"Y^ to now", "2025-01-01T00:00:00Z", "2025-03-12T10:32:36.001Z", timeutil.TimeGrainSecond},
		{"-1Y!", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainSecond},
		{"W1 of Y", "2024-12-30T00:00:00Z", "2025-01-06T00:00:00Z", timeutil.TimeGrainSecond},
		{"W1 of -1M^ to -1M$", "2025-02-03T00:00:00Z", "2025-02-10T00:00:00Z", timeutil.TimeGrainSecond},
		{"-2d^ to d$ as of -1Q", "2024-12-08T00:00:00Z", "2024-12-11T00:00:00Z", timeutil.TimeGrainSecond},

		{"<6H of D25 as of -3M", "2024-12-25T00:00:00Z", "2024-12-25T06:00:00Z", timeutil.TimeGrainSecond},
		{"6h starting D25^ as of -3M", "2024-12-25T00:00:00Z", "2024-12-25T06:00:00Z", timeutil.TimeGrainSecond},

		{"-4d^ to now", "2025-03-06T00:00:00Z", "2025-03-12T10:32:36.001Z", timeutil.TimeGrainSecond},
		{"M/MW^ to M/MW^+3W", "2025-03-03T00:00:00Z", "2025-03-24T00:00:00Z", timeutil.TimeGrainSecond},
		{"3W starting M^", "2025-03-03T00:00:00Z", "2025-03-24T00:00:00Z", timeutil.TimeGrainSecond},
		{"1Y starting Y^", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainSecond},

		{">7h of -1d!", "2025-03-09T17:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainSecond},
		{"H4 as of -1d", "2025-03-09T03:00:00Z", "2025-03-09T04:00:00Z", timeutil.TimeGrainSecond},
		{"H4 of -1d!", "2025-03-09T03:00:00Z", "2025-03-09T04:00:00Z", timeutil.TimeGrainSecond},
		{"3d ending -1Q/d$", "2024-12-08T00:00:00Z", "2024-12-11T00:00:00Z", timeutil.TimeGrainSecond},

		{"y/yw^ to w^", "2024-12-30T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainSecond},
		{"m30 of H12 of D5 of >1W of Q3", "2025-09-26T11:29:00Z", "2025-09-26T11:30:00Z", timeutil.TimeGrainSecond},
		{"m30 of H12 of D5 of >1W of Q3 as of -2Y", "2023-09-29T11:29:00Z", "2023-09-29T11:30:00Z", timeutil.TimeGrainSecond},
		{"m30 of H12 of D5 of >1W of Q3 as of -4Y", "2021-10-01T11:29:00Z", "2021-10-01T11:30:00Z", timeutil.TimeGrainSecond},
		{"m30 of H12 of D5 of >1W of Q3 as of -5Y", "2020-09-25T11:29:00Z", "2020-09-25T11:30:00Z", timeutil.TimeGrainSecond},
		{"-5W4M3Q2Y to -4W3M2Q1Y", "2022-01-06T06:32:36Z", "2023-05-13T06:32:36Z", timeutil.TimeGrainSecond},
		{"-5W-4M-3Q-2Y to -4W-3M-2Q-1Y", "2022-01-06T06:32:36Z", "2023-05-13T06:32:36Z", timeutil.TimeGrainSecond},
	}

	for _, testCase := range testCases {
		t.Run(testCase.timeRange, func(t *testing.T) {
			rt, err := Parse(testCase.timeRange, ParseOptions{})
			require.NoError(t, err)

			start, end, grain := rt.Eval(EvalOptions{
				Now:       now,
				MinTime:   minTime,
				MaxTime:   maxTime,
				Watermark: watermark,
			})
			fmt.Println(start, end, grain)
			require.Equal(t, parseTestTime(t, testCase.start), start)
			require.Equal(t, parseTestTime(t, testCase.end), end)
			require.Equal(t, testCase.grain, grain)
		})
	}
}

func Test_Eval_watermark_on_boundary(t *testing.T) {
	now := "2026-05-14T10:32:36Z"
	minTime := "2020-01-01T00:32:36Z"
	maxTime := "2025-07-01T00:00:00Z"   // month and quarter boundary
	watermark := "2025-05-12T00:00:00Z" // day and week boundary
	testCases := []testCase{
		{"-2D^ to D^", "2025-05-10T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay},
		{"-2D$ to D$", "2025-05-11T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay},
		{"-4D^ to -2D^", "2025-05-08T00:00:00Z", "2025-05-10T00:00:00Z", timeutil.TimeGrainDay},
		{"-1D^ to D^", "2025-05-11T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainHour},
		{"D^ to D$", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour},

		{"-2D!", "2025-05-10T00:00:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainHour},
		{"-2W!", "2025-04-28T00:00:00Z", "2025-05-05T00:00:00Z", timeutil.TimeGrainDay},
		{"-2M! as of latest", "2025-05-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainDay},
		{"-2Q! as of latest", "2025-01-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainMonth},

		{"H2 of -1D!", "2025-05-11T01:00:00Z", "2025-05-11T02:00:00Z", timeutil.TimeGrainMinute},
		{"D2 of -1W!", "2025-05-06T00:00:00Z", "2025-05-07T00:00:00Z", timeutil.TimeGrainHour},
		{"W2 of -1M! as of latest", "2025-06-09T00:00:00Z", "2025-06-16T00:00:00Z", timeutil.TimeGrainDay},
		{"W2 of -1Q! as of latest", "2025-04-07T00:00:00Z", "2025-04-14T00:00:00Z", timeutil.TimeGrainDay},
		{"W2 of -1Y! as of 2024", "2023-01-09T00:00:00Z", "2023-01-16T00:00:00Z", timeutil.TimeGrainDay},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

type testCase struct {
	timeRange string
	start     string
	end       string
	grain     timeutil.TimeGrain
}

type testGrainPair struct {
	grain, higherGrain                              string
	lastStart, higherPeriod, ordinalStart, firstEnd string
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
				Now:       nowTm,
				MinTime:   minTimeTm,
				MaxTime:   maxTimeTm,
				Watermark: watermarkTm,
			})
			fmt.Println(start, end, grain)
			require.Equal(t, parseTestTime(t, testCase.start), start)
			require.Equal(t, parseTestTime(t, testCase.end), end)
			require.Equal(t, testCase.grain, grain)
		})
	}
}

func parseTestTime(tst *testing.T, t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return ts
}
