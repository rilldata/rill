---
title: "All About Time"
sidebar_label: "Time Series Syntax"
hide_table_of_contents: false
tags:
    - Quickstart
    - Tutorial
---

In this guide, we will discuss the fine details of our time series syntax, what each parameter means and has the capability doing, and how to create your own time ranges as both a developer and a consumer.

## What is time range syntax?

Basically, we found that many of our users needed a more robust systme to be able to include/exclude certain area of their KPI due to refresh schedules. The biggest feedback we heard is that if the data is refreshed daily, certain time zones will see a huge drop in their measures due to incomplete data being shown. This raised a lot of questions to the team.

With our current time range syntax, a user and/or developer can create their own time ranges based on the needs of their team. Let's start the parameters that you can set in your metrics views.


## Metrics View Parameters

- ***`timeseries`***: -
- ***`watermark`***: -
- ***`smallest_time_grain`***: -
- ***`first_day_of_week`***: -
- ***`first_month_of_year`***: -

## Understanding the syntax

Now that we understand some of the underlying components to the syntax, let's discuss how the syntax works.

### Time Periods

There are a few different types of time ranges
1. Time series between two fixed time periods (not current time)
2. Time series between two fixed time periods (current time)
3. ISO 8601 time range

You'll see why we separated two fixed time periods into two sections later.

### Grain


### `as of ...`

## Creating your own time ranges

### ISO 8601 Time Ranges

```
```

### Time Series Between Two Fixed Points (current time)

```
```

### Time Series Between Two Fixed Points (not to current time)
```
```




