---
title: Uplift Calculation Using Required Fields Example Metric
tags:
- metrics
- code
- snippets
- clickhouse
docs: https://docs.rilldata.com/build/metrics-view/measures/referencing
hash: 7fecf603decc5fc1a5d282927fd75344014aef7cf9d7bb3686badb4bd20e3543
---

```yaml
- name: uplift_percent
  display_name: "% Uplift in Prebid"
  description: Value per Impression, Active vs Control
  format_preset: percentage
  requires:
    [
      router_value,
      controller_value,
      router_impressions,
      controller_impressions
    ]
  expression: |
    (
      (router_value/router_impressions - controller_value/controller_impressions)
      /
      (controller_value/controller_impressions)
    )
  valid_percent_of_total: false
# Intermediate calculations
- name: router_impressions
  display_name: Enriched Cohort
  expression: sumIf(__sourcerows, gctest)
  description: Total ENRICHED Impressions
  format_preset: none
  valid_percent_of_total: true

- name: controller_impressions
  display_name: Control Cohort
  expression: sumIf(__sourcerows, NOT gctest)
  description: Total number of CONTROL impressions (un-ENRICHED)
  format_preset: none
  valid_percent_of_total: false

- name: router_value
  display_name: Enriched Cohort Value
  expression: sumIf(normalized_value, gctest)
  description: Total value of Active cohort
  format_preset: none
  valid_percent_of_total: false

- name: controller_value
  display_name: Control Cohort Value
  expression: sumIf(normalized_value, NOT gctest)
  description: Total value of Test cohort
  format_preset: none
  valid_percent_of_total: false
```
