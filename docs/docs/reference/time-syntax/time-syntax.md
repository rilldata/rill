---
title: Time Range Syntax
description: Complete reference for Rill's time range syntax
sidebar_label: Time Range Syntax
sidebar_position: 0
---

Rill provides a flexible, expressive syntax for specifying time ranges. This reference covers every aspect of the syntax with detailed examples.

## Grammar Overview

```
expression     = interval [as_of]* [by_clause] [tz_clause] [offset_clause]

interval       = shorthand | period_to_date | start_end | ordinal | iso_interval

shorthand      = duration
period_to_date = grain "TD"
start_end      = point_in_time "to" point_in_time
ordinal        = grain number ["of" grain number]*
iso_interval   = iso_timestamp ["to" iso_timestamp]

point_in_time  = (offset | reference) [snap]*
offset         = prefix duration
reference      = "ref" | "now" | "watermark" | "latest" | "earliest"
snap           = "/" grain
duration       = (number grain)+
prefix         = "+" | "-"

as_of          = "as" "of" point_in_time
by_clause      = "by" grain
tz_clause      = "tz" timezone
offset_clause  = "offset" (prefix number "P" | prefix duration)

grain          = "s" | "m" | "h" | "D" | "W" | "M" | "Q" | "Y"
```

---

## Time Grains

| Grain | Aliases | Description | Examples |
|-------|---------|-------------|----------|
| `s` | `S` | Second | `30s`, `sTD` |
| `m` | — | Minute | `15m`, `mTD` |
| `h` | `H` | Hour | `24h`, `hTD`, `HTD` |
| `d` | `D` | Day | `7D`, `7d`, `DTD` |
| `w` | `W` | Week | `2W`, `WTD` |
| `M` | — | Month | `3M`, `MTD` |
| `q` | `Q` | Quarter | `1Q`, `QTD` |
| `y` | `Y` | Year | `1Y`, `YTD` |

:::warning Month vs Minute
`M` is always **month**. `m` is always **minute**. This is the only case-sensitive distinction.
:::

---

## Reference Points

| Keyword | Description | Use Case |
|---------|-------------|----------|
| `ref` | Contextual reference time | Internal use, modified by `as of` |
| `now` | Current wallclock time | Real-time dashboards |
| `watermark` | Data completeness marker | Production dashboards with ETL delays |
| `latest` | Most recent data timestamp | Show data up to latest available |
| `earliest` | Oldest data timestamp | Full historical analysis |

### Reference Point Examples

Given: `now = 2025-05-15T10:32:36Z`, `watermark = 2025-05-13T06:32:36Z`, `latest = 2025-05-14T06:32:36Z`

| Expression | Start | End |
|------------|-------|-----|
| `7D` | `2025-05-08T10:32:36Z` | `2025-05-15T10:32:36Z` |
| `7D as of now` | `2025-05-08T10:32:36Z` | `2025-05-15T10:32:36Z` |
| `7D as of watermark` | `2025-05-06T06:32:36Z` | `2025-05-13T06:32:36Z` |
| `7D as of now/D` | `2025-05-08T00:00:00Z` | `2025-05-15T00:00:00Z` |
| `7D as of watermark/H+1H` | `2025-05-06T07:00:00Z` | `2025-05-13T07:00:00Z` |

---

## Interval Types

### 1. Shorthand Intervals

The simplest form: a duration that counts backward from the reference point.

**Syntax:** `<number><grain>[<number><grain>]*`

**Expansion:** `<duration>` → `-<duration> to ref`

| Shorthand | Equivalent Start-End | Description |
|-----------|---------------------|-------------|
| `7D` | `-7D to ref` | Last 7 days |
| `4h` | `-4h to ref` | Last 4 hours |
| `2W` | `-2W to ref` | Last 2 weeks |
| `3M` | `-3M to ref` | Last 3 months |
| `1Y` | `-1Y to ref` | Last 1 year |

**Multi-grain durations:**

| Expression | Description |
|------------|-------------|
| `3W18D23h` | 3 weeks + 18 days + 23 hours |
| `1Y6M` | 1 year + 6 months |
| `2D12h30m` | 2 days + 12 hours + 30 minutes |

**Examples with reference points:**

Given `watermark = 2025-05-13T06:32:36Z`:

| Expression | Start | End |
|------------|-------|-----|
| `7D as of watermark` | `2025-05-06T06:32:36Z` | `2025-05-13T06:32:36Z` |
| `MTD as of watermark` | `2025-05-01T00:00:00Z` | `2025-05-13T06:32:36Z` |

---

### 2. Period-to-Date Intervals

Returns from the start of the current period to the reference point.

**Syntax:** `<grain>TD`

**Expansion:** `<grain>TD` → `ref/<grain> to ref`

| Syntax | Expansion | Description |
|--------|-----------|-------------|
| `sTD` | `ref/s to ref` | Current second |
| `mTD` | `ref/m to ref` | Minute to date |
| `hTD` | `ref/h to ref` | Hour to date |
| `DTD` | `ref/D to ref` | Day to date (today so far) |
| `WTD` | `ref/W to ref` | Week to date |
| `MTD` | `ref/M to ref` | Month to date |
| `QTD` | `ref/Q to ref` | Quarter to date |
| `YTD` | `ref/Y to ref` | Year to date |

**Examples:**

Given `now = 2025-05-15T10:32:36Z`:

| Expression | Start | End |
|------------|-------|-----|
| `DTD` | `2025-05-15T00:00:00Z` | `2025-05-15T10:32:36Z` |
| `MTD` | `2025-05-01T00:00:00Z` | `2025-05-15T10:32:36Z` |
| `YTD` | `2025-01-01T00:00:00Z` | `2025-05-15T10:32:36Z` |
| `MTD as of watermark` | `2025-05-01T00:00:00Z` | `2025-05-13T06:32:36Z` |

---

### 3. Start-End Intervals

Explicit specification of start and end points.

**Syntax:** `<point_in_time> to <point_in_time>`

#### Point-in-Time Components

A point in time can be:
- **Reference:** `ref`, `now`, `watermark`, `latest`, `earliest`
- **Offset:** `-4D`, `+1W`, `-2M3D`
- **Snapped:** `now/D`, `watermark/W`, `-4D/D`
- **Combined:** `-4D/D+2h`, `watermark/M-1D`

#### Detailed Examples

Given `now = 2025-05-15T10:32:36Z`, `watermark = 2025-05-13T06:32:36Z`:

| Expression | Start | End | Notes |
|------------|-------|-----|-------|
| `-4d to now` | `2025-05-11T10:32:36Z` | `2025-05-15T10:32:36Z` | Exact 4 days ago |
| `-4d/d to now/d` | `2025-05-11T00:00:00Z` | `2025-05-15T00:00:00Z` | Snapped to day boundaries |
| `-4d to now/d` | `2025-05-11T10:32:36Z` | `2025-05-15T00:00:00Z` | Mixed: exact start, snapped end |
| `watermark/D to watermark` | `2025-05-13T00:00:00Z` | `2025-05-13T06:32:36Z` | Start of watermark day to watermark |
| `watermark/Y to watermark` | `2025-01-01T00:00:00Z` | `2025-05-13T06:32:36Z` | Year to watermark |
| `watermark to latest` | `2025-05-13T06:32:36Z` | `2025-05-14T06:32:36Z` | Between watermark and latest |
| `earliest to latest` | *(earliest data)* | *(latest data)* | All data |

#### First/Last N of a Period

| Expression | Description | Example Result |
|------------|-------------|----------------|
| `-2m/m-2s to -2m/m` | Last 2 seconds of 2 mins ago | `06:29:58Z` to `06:30:00Z` |
| `-2m/m to -2m/m+2s` | First 2 seconds of 2 mins ago | `06:30:00Z` to `06:30:02Z` |
| `-2h/h-2m to -2h/h` | Last 2 minutes of 2 hours ago | `03:58:00Z` to `04:00:00Z` |
| `-2h/h to -2h/h+2m` | First 2 minutes of 2 hours ago | `04:00:00Z` to `04:02:00Z` |
| `-2D/D-2h to -2D/D` | Last 2 hours of 2 days ago | `22:00:00Z` to `00:00:00Z` |
| `-2D/D to -2D/D+2h` | First 2 hours of 2 days ago | `00:00:00Z` to `02:00:00Z` |
| `-2W/W-2D to -2W/W` | Last 2 days of 2 weeks ago | 2 days before week boundary |
| `-2W/W to -2W/W+2D` | First 2 days of 2 weeks ago | First 2 days of that week |
| `-2M/M-2D to -2M/M` | Last 2 days of 2 months ago | Last 2 days of that month |
| `-2M/M to -2M/M+2D` | First 2 days of 2 months ago | First 2 days of that month |
| `-2Q/Q-2M to -2Q/Q` | Last 2 months of 2 quarters ago | Last 2 months of that quarter |
| `-2Y/Y-2M to -2Y/Y` | Last 2 months of 2 years ago | Nov-Dec of that year |
| `-2Y/Y-2Q to -2Y/Y` | Last 2 quarters of 2 years ago | Q3-Q4 of that year |

---

### 4. Ordinal Intervals

Selects a specific numbered occurrence within a period.

**Syntax:** `<grain><number> [of <grain><number>]*`

#### Basic Ordinals

| Expression | Description |
|------------|-------------|
| `D1` | 1st day of reference period |
| `D15` | 15th day of reference period |
| `W1` | 1st week of reference period |
| `W2` | 2nd week of reference period |
| `M1` | 1st month (January) |
| `M11` | 11th month (November) |
| `Q2` | 2nd quarter (Apr-Jun) |
| `H12` | 12th hour (noon) |
| `m30` | 30th minute |
| `s45` | 45th second |

#### Chained Ordinals

| Expression | Description |
|------------|-------------|
| `D3 of M2` | 3rd day of February |
| `W2 of Q1` | 2nd week of Q1 |
| `D15 of M6` | June 15th |
| `H2 of D4` | 2nd hour of 4th day |
| `m15 of H2` | 15th minute of 2nd hour |
| `s30 of m15 of H2` | 30th second of 15th minute of 2nd hour |
| `s57 of m4 of H2 of D4` | Very specific moment in 4th day |

#### Ordinals with Reference Points

Given `watermark = 2025-05-13T06:32:36Z`:

| Expression | Start | End |
|------------|-------|-----|
| `W1` | `2025-04-28T00:00:00Z` | `2025-05-05T00:00:00Z` |
| `W1 as of -2M` | `2025-03-03T00:00:00Z` | `2025-03-10T00:00:00Z` |
| `D2 as of watermark/W` | `2025-05-13T00:00:00Z` | `2025-05-14T00:00:00Z` |
| `D2 as of -1W as of watermark/W` | `2025-05-06T00:00:00Z` | `2025-05-07T00:00:00Z` |
| `W2 as of -1M as of latest/M` | `2025-06-09T00:00:00Z` | `2025-06-16T00:00:00Z` |
| `W2 as of -1Q as of latest/Q` | `2025-04-07T00:00:00Z` | `2025-04-14T00:00:00Z` |
| `W2 as of -1Y as of 2024` | `2023-01-09T00:00:00Z` | `2023-01-16T00:00:00Z` |
| `M2 as of -2Q/Q as of watermark/Q` | `2024-11-01T00:00:00Z` | `2024-12-01T00:00:00Z` |
| `Q2 as of -2Y/Y as of watermark/Y` | `2023-04-01T00:00:00Z` | `2023-07-01T00:00:00Z` |

#### Complex Ordinal Examples

| Expression | Description |
|------------|-------------|
| `s57 of m4 of H2 of D4 as of -1M` | 57th sec of 4th min of 2nd hour of 4th day of last month |
| `D3 of M11 as of 2024` | November 3rd, 2024 |
| `W2 as of -2M/M as of watermark/M` | 2nd week of 2 months ago |

---

### 5. ISO 8601 Intervals

Standard date/time format for absolute time specifications.

#### Single Timestamps (Implicit Range)

A single timestamp expands to cover its entire grain:

| Expression | Start | End | Grain |
|------------|-------|-----|-------|
| `2025` | `2025-01-01T00:00:00Z` | `2026-01-01T00:00:00Z` | Year |
| `2025-02` | `2025-02-01T00:00:00Z` | `2025-03-01T00:00:00Z` | Month |
| `2025-02-20` | `2025-02-20T00:00:00Z` | `2025-02-21T00:00:00Z` | Day |
| `2025-02-20T01` | `2025-02-20T01:00:00Z` | `2025-02-20T02:00:00Z` | Hour |
| `2025-02-20T01:23` | `2025-02-20T01:23:00Z` | `2025-02-20T01:24:00Z` | Minute |
| `2025-02-20T01:23:45Z` | `2025-02-20T01:23:45Z` | `2025-02-20T01:23:46Z` | Second |

#### Explicit Ranges

Multiple separators are supported: `to`, `/`, `,`

| Expression | Start | End |
|------------|-------|-----|
| `2025-02-20T01:23:45Z to 2025-07-15T02:34:50Z` | `2025-02-20T01:23:45Z` | `2025-07-15T02:34:50Z` |
| `2025-02-20T01:23:45Z / 2025-07-15T02:34:50Z` | `2025-02-20T01:23:45Z` | `2025-07-15T02:34:50Z` |
| `2025-02-20T01:23:45Z,2025-07-15T02:34:50Z` | `2025-02-20T01:23:45Z` | `2025-07-15T02:34:50Z` |

#### Sub-second Precision

| Expression | Precision |
|------------|-----------|
| `2025-02-20T01:23:45.123Z` | Milliseconds |
| `2025-02-20T01:23:45.123456Z` | Microseconds |
| `2025-02-20T01:23:45.123456789Z` | Nanoseconds |

---

## Snapping (Boundary Alignment)

Snapping truncates a time to a grain boundary using the `/` operator.

**Syntax:** `<point_in_time>/<grain>`

### Snap Behavior

| Expression | Given `now = 2025-05-15T10:32:36Z` | Result |
|------------|-----------------------------------|--------|
| `now/s` | Snap to second | `2025-05-15T10:32:36Z` |
| `now/m` | Snap to minute | `2025-05-15T10:32:00Z` |
| `now/h` | Snap to hour | `2025-05-15T10:00:00Z` |
| `now/D` | Snap to day | `2025-05-15T00:00:00Z` |
| `now/W` | Snap to week (Monday) | `2025-05-12T00:00:00Z` |
| `now/M` | Snap to month | `2025-05-01T00:00:00Z` |
| `now/Q` | Snap to quarter | `2025-04-01T00:00:00Z` |
| `now/Y` | Snap to year | `2025-01-01T00:00:00Z` |

### Double Snapping for ISO Week Boundaries

Weeks can span year boundaries (ISO 8601 week rules). Double snapping handles this:

| Expression | Description |
|------------|-------------|
| `-2Y/Y/W to -1Y/Y/W` | First week of 2 years ago to first week of last year |
| `-0Y/Y/W to ref/W` | First week of this year to current week |

The second `/W` applies ISO week correction to align properly with week boundaries that may fall in the previous/next year.

**Example:**
```
-2Y/Y/W to -1Y/Y/W as of watermark
→ 2023-01-02T00:00:00Z to 2024-01-01T00:00:00Z
```

---

## Modifiers

### `as of` — Reference Point Override

Changes the reference point for time calculations. Multiple `as of` clauses are evaluated right-to-left.

**Syntax:** `<expression> as of <point_in_time>`

| Expression | Description |
|------------|-------------|
| `7D as of now` | 7 days ending at current time |
| `7D as of now/D` | 7 days ending at start of today |
| `7D as of watermark` | 7 days ending at watermark |
| `7D as of watermark/D` | 7 days ending at start of watermark day |
| `7D as of watermark/D+1D` | 7 days ending at start of tomorrow (includes today) |
| `7D as of latest/D+1D` | 7 days ending at day after latest data |

**Chaining `as of`:**

| Expression | Evaluation Order |
|------------|------------------|
| `D3 as of -1M as of watermark/M` | `watermark/M` → `-1M` → `D3` |
| `W2 as of -1Y as of 2024` | `2024` → `-1Y` → `W2` |
| `-2D/D to ref/D as of -2D as of watermark/D+1D` | Complex comparison setup |

### `by` — Display Grain

Specifies the aggregation grain for display purposes.

**Syntax:** `<expression> by <grain>`

| Expression | Description |
|------------|-------------|
| `7D by D` | Last 7 days, aggregate by day |
| `MTD by h` | Month to date, aggregate by hour |
| `YTD by M` | Year to date, aggregate by month |
| `1Q by W` | Last quarter, aggregate by week |

### `tz` — Timezone

Specifies the timezone for evaluation. Uses IANA timezone names.

**Syntax:** `<expression> tz <timezone>`

| Expression | Description |
|------------|-------------|
| `7D tz America/New_York` | 7 days in Eastern time |
| `MTD tz Europe/London` | Month to date in UK time |
| `DTD tz Asia/Tokyo` | Today in Japan time |
| `W1 as of watermark tz Asia/Kathmandu` | First week in Nepal time |

**Daylight Saving Time Examples (America/New_York):**

| Expression | Notes |
|------------|-------|
| `D3 of M11 as of 2024` | Nov 3, 2024 (DST ends) |
| `D10 of M3 as of 2024` | Mar 10, 2024 (DST begins) |
| `3D as of 2024-11-04` | Crosses DST boundary |
| `2M as of 2024-12` | Includes DST transition |

### `offset` — Time Shift for Comparisons

Shifts the entire time range for comparison purposes.

**Syntax:** `<expression> offset <shift>`

#### Grain-based Offset

| Expression | Description |
|------------|-------------|
| `7D offset -1W` | Previous week's 7-day period |
| `7D offset -1M` | Same 7 days, one month earlier |
| `MTD offset -1Y` | Same MTD, one year ago |
| `7D as of latest/D+1D offset -1M` | Last 7 days shifted back 1 month |

#### Previous Period Offset

| Expression | Description |
|------------|-------------|
| `7D offset -1P` | The 7 days before the current 7 days |
| `MTD offset -1P` | Previous month's MTD equivalent |
| `2025-02-20 offset -1P` | The day before (Feb 19) |
| `2025-02 offset -1P` | Previous month (January) |
| `2025 offset -1P` | Previous year (2024) |

**ISO Range with Previous Period:**

| Expression | Start | End |
|------------|-------|-----|
| `2025-02-20T01:23:45Z,2025-07-15T02:34:50Z offset -1P` | `2024-09-28T00:12:40Z` | `2025-02-20T01:23:45Z` |

---

## Special Keywords

### `inf` — All Time

Returns all available data from earliest to latest.

```
inf → earliest to latest/s+1s
```

**Example:**
Given `earliest = 2020-01-01T00:32:36Z`, `latest = 2025-05-14T06:32:36Z`:
```
inf → 2020-01-01T00:32:36Z to 2025-05-14T06:32:37Z
```

---

## Complete and Incomplete Periods

### Excluding Current (Incomplete) Period

| Goal | Expression |
|------|------------|
| Last 7 complete days | `7D as of watermark/D` |
| Previous complete week | `1W as of watermark/W` |
| Previous complete month | `1M as of watermark/M` |
| Previous complete quarter | `1Q as of watermark/Q` |
| Previous complete year | `1Y as of watermark/Y` |

### Including Current (Incomplete) Period

| Goal | Expression |
|------|------------|
| Last 7 days including today | `7D as of watermark/D+1D` |
| Current week including today | `1W as of watermark/W+1W` |
| Current month including today | `1M as of watermark/M+1M` |
| Current quarter including today | `1Q as of watermark/Q+1Q` |
| Current year including today | `1Y as of watermark/Y+1Y` |

### Watermark Boundary Examples

Given `watermark = 2025-05-12T00:00:00Z` (exactly on day/week boundary):

| Expression | Start | End |
|------------|-------|-----|
| `1h as of watermark/h` | `2025-05-11T23:00:00Z` | `2025-05-12T00:00:00Z` |
| `1h as of watermark/h+1h` | `2025-05-12T00:00:00Z` | `2025-05-12T01:00:00Z` |
| `1D as of watermark/D` | `2025-05-11T00:00:00Z` | `2025-05-12T00:00:00Z` |
| `1D as of watermark/D+1D` | `2025-05-12T00:00:00Z` | `2025-05-13T00:00:00Z` |
| `2D as of watermark/D` | `2025-05-10T00:00:00Z` | `2025-05-12T00:00:00Z` |
| `2D as of watermark/D+1D` | `2025-05-11T00:00:00Z` | `2025-05-13T00:00:00Z` |
| `2W as of watermark/W` | `2025-04-28T00:00:00Z` | `2025-05-12T00:00:00Z` |

---

## Duration vs Fixed Range

A critical distinction that often causes confusion:

### The Problem

Given `now = 2025-09-03T14:30:00Z`:

| Expression | Start | End | Duration |
|------------|-------|-----|----------|
| `-4d to now/d` | `2025-08-30T14:30:00Z` | `2025-09-03T00:00:00Z` | ~3.4 days |
| `4d as of now/d` | `2025-08-30T00:00:00Z` | `2025-09-03T00:00:00Z` | Exactly 4 days |

### Why They Differ

- **`-4d to now/d`**: Start is calculated from wallclock (2:30 PM), end is snapped to midnight
- **`4d as of now/d`**: Reference is snapped first, then 4 days calculated backward

### Best Practices

| Goal | Recommended | Avoid |
|------|-------------|-------|
| Exact N days | `Nd as of ref/D` | `-Nd to ref/D` |
| Exact N weeks | `Nw as of ref/W` | `-Nw to ref/W` |
| Partial periods OK | `-Nd to ref` | — |

---

## Week Correction (ISO 8601)

Rill follows ISO 8601 week rules where:
- Weeks start on Monday
- Week 1 contains the first Thursday of the year
- Weeks can span year boundaries

### First Week Variations by Boundary Day

| Boundary Day | Week Starts Monday | Week Starts Sunday |
|--------------|-------------------|-------------------|
| Monday | Same day | Previous day |
| Tuesday | Previous day | 2 days before |
| Wednesday | 2 days before | 3 days before |
| Thursday | 3 days before | Next week |
| Friday | Next week | Next week |
| Saturday | Next week | Next week |
| Sunday | Next week | Same day |

**Example:** `W1 as of 2025-01-01T00:00:00Z` (Wednesday)
- Week starts Monday: `2024-12-30T00:00:00Z` to `2025-01-06T00:00:00Z`
- Week starts Sunday: `2024-12-29T00:00:00Z` to `2025-01-05T00:00:00Z`
