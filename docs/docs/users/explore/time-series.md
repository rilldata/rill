---
title: "Time Series Filter"
sidebar_label: "Time Series Filter"
hide_table_of_contents: false
sidebar_position: 15
tags:
    - Quickstart
    - Tutorial
---

Once you've [built your metrics view](/build/metrics-view) and assigned a time series column, you'll be able to start visualizing your measures and dimensions in either an Explore dashboard or Canvas dashboard. This guide will discuss all the features in the time navigator that exists at the top of all dashboards and how to customize it to your needs.

## Time Series Filter Component

The time series navigation is divided into four main sections:

### 1. Forward/Backward Navigation
Allows you to step forward or backward through your selected time range. Use the arrow buttons to navigate through time periods while maintaining your current selection.

### 2. Time Selector
The dropdown menu containing default time ranges (like "Last 7 days", "Last 30 days") or custom time ranges you've created. This is where you choose your primary time period. It also has a selector for custom ranges in a calendar view as well as time zone selector.

### 3. As Of...
Controls the reference point for your time calculations. This determines whether your time range is relative to the current moment (wallclock), latest data based on grain, or "completed" data.

### 4. Comparing
Enables time comparison functionality, allowing you to compare current data against previous periods (e.g., "vs. previous period" or "vs. same period last year").

<img src='/img/build/metrics-view/time-series/time-pill.png' class='rounded-gif' />
<br />

## Key Concepts

Before diving into this complicated topic, it's important to understand these fundamental concepts:

- **Reference Points**: The anchor point from which durations are calculated
- **Snapping**: Aligning time boundaries to specific grains (day, hour, etc.)
- **Duration vs Fixed Points**: The difference between "4 days ago" and "4 days ending at a specific point"
- **Wallclock Time**: The actual current time vs. data-relative time



### Time Selection Types

There are several ways to specify time ranges:

1. **Duration-based ranges**: Specify a length of time from a reference point (e.g., `7d as of latest/d`)
2. **Fixed-point ranges**: Specify exact start and end points (e.g., `-4d to now/d`)
3. **ISO 8601 ranges**: Use standard date format (e.g., `2024-01-01 to 2024-01-05`)

The following sections explain how each type works and when to use them.

## Time Ranges Explained

Understanding the difference between duration-based and fixed-point ranges is crucial for avoiding confusion.

### Duration-Based Ranges
These specify a length of time from a reference point:
- `4d` = 4 days duration
- `4d as of now/d` = 4 days ending at start of current day

### Fixed-Point Ranges  
These specify exact start and end points:
- `-4d to now/d` = from 4 days ago (wallclock) to start of current day
- `2024-01-01 to 2024-01-05` = exact date range

:::tip Not the same

Many users expect `-4d to now/d` and `4d as of now/d` to be equivalent, but they're not:

- **`-4d to now/d`**: Starts 4 days ago from current wallclock time, ends at start of current day
  - If it's 12:30 PM on Sept 3, this starts at 12:30 PM on Aug 30
  - Duration: ~3.5 days (partial Aug 30 + full Aug 31, Sep 1, Sep 2)

- **`4d as of now/d`**: Exactly 4 days ending at start of current day  
  - Starts at start of Aug 30, ends at start of Sept 3
  - Duration: exactly 4 full days

:::

### ISO 8601 Ranges

ISO 8601 ranges use the standard date format for specifying exact time periods:

- `2024-01-01 to 2024-01-05` - From January 1st to January 5th, 2024
- `2024-01-01T00:00:00Z to 2024-01-01T23:59:59Z` - Full day with explicit timestamps
- `2024-01-01T09:00:00 to 2024-01-01T17:00:00` - Business hours on a specific day

These ranges are useful when you need precise control over the exact start and end times, especially for reporting on specific events or periods.

### Snapping 

Snapping aligns time boundaries to specific grains (day, hour, etc.) rather than using exact wallclock times. This allows your dashboards to show complete summaries of data rather than incomplete sets.

- `/s` = snap to second boundaries
- `/m` = snap to minute boundaries
- `/h` = snap to hour boundaries  
- `/d` = snap to day boundaries (start of day)
- `/w` = snap to week boundaries (start of week)
- `/M` = snap to month boundaries (start of month)
- `/y` = snap to year boundaries (start of year)


## As of
### Reference
- **Complete data `watermark`**: Uses the [watermark timestamp](/build/metrics-view/time-series#watermark) from your metrics view. This ensures you only see data that has been fully processed and is considered "complete" according to your data pipeline's watermark settings.
- **Latest data `latest`**: Uses the most recent data point available in your dataset, regardless of completeness. This is useful when you want to see the freshest data even if it might be incomplete.
- **Current time `now`**: Uses the current wallclock time as the reference point. This means your time ranges will always be relative to the present moment, which can include future time periods if your data extends beyond the current time.

:::tip Unsure which one?

If you hover on any of the three options, it will give you the actual time that will be considered in the time filter.

:::

### Grain

The grain determines the time granularity for your reference point. This affects how your time ranges can be rolled up and displayed. Depending on the `smallest_time_grain` in your metrics view, the options will be limited.

You can also configure grain settings directly in the Time Selector using the [snapping](#snapping) options.

### Anchor

The anchor determines whether to include incomplete time periods in your data. After snapping to a specific grain, you can choose to:

- **Include incomplete periods**: Show data for the current partial time grain (e.g., today's data even if the day isn't finished)
- **Exclude incomplete periods**: Only show data for complete time grains (e.g., exclude today if it's not finished)

This is similar to the `watermark` concept but operates at the grain level you've selected, providing more granular control over data completeness. 


## Bringing it all together

The following are the most common sources of confusion and how to avoid them:

### 1. "I want the last 4 calendar days"

**Confusing approaches:**
- `-4d to now/d` - This gives you ~3.5 days if it's not midnight
- `4d` - This gives you exactly 4 days from current wallclock time

**Correct approaches:**
- `4d as of now/d` - Exactly 4 days ending at start of current day
- `-4d/d to now/d` - From start of day 4 days ago to start of current day
- `-4d to ref as of now/d` - 4 days ending at start of current day

### 2. "Why does my range show 3 days when I asked for 4?"

This happens when you mix wallclock time with snapped boundaries:
- `-4d to now/d` at 12:30 PM = Aug 30 12:30 PM to Sept 3 00:00 AM
- Duration: Partial Aug 30 + Aug 31 + Sep 1 + Sep 2 = ~3.5 days

**Solution:** Use consistent snapping: `-4d/d to now/d`

### 3. "The end date keeps shifting when I change the 'as of'"

This happens because the end date uses `now` (wallclock time) instead of a snapped reference:
- `-4d to now/d` - End date is always start of current day
- `-4d to now/d as of now/d+1d` - End date becomes start of tomorrow

**Solution:** Use duration-based ranges: `4d as of now/d`

### 4. "I want to look at data relative to the latest data point"

**Problem:** Using `now` when data might be stale
- `-7d as of latest/d` - 7 days ending at latest data day
- `-7d` - 7 days ending at current wallclock time (might include future)

**Solution:** Always specify the reference point explicitly

## Time Comparisons

Along with setting time ranges, you have the ability to set a comparison period in the "Comparing" section of the time pill. This allows you to analyze trends and changes by comparing your current time period against a previous period.

### Comparison Types

1. **Previous Period**: Compares against the immediately preceding period of the same duration
2. **Previous ...**: Compares current selected period vs a set period (day, week, month, year)
3. **Custom Comparison**: Set a specific comparison period using the same time range syntax

### Understanding Comparison Results

Once comparison is enabled, you'll see slightly different information in your dashboard. Along with the current value, you'll see both change in value as well as % change over periods. This gives you a quick glance at how your metrics are performing compared to the previous period.

<img src = '/img/explore/filters/comparison.png' class='rounded-gif' />
<br />

## Filter by Scrubbing 

For a specific view into your time series graph, you can interactively scrub directly on the time series graph. This feature allows you to zoom into specific time periods by clicking and dragging across the chart.

<img src = '/img/explore/filters/scrub.png' class='rounded-gif' />
<br />
