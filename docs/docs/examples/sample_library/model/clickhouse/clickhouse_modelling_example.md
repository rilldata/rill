---
title: Example file for Rill modelling on ClickHouse
tags:
- model
- code
- complete_file
- clickhouse
docs: https://docs.rilldata.com/build/models
hash: 10a8b52141c6aabb825984891be18e3e2c0bd8bf7dd98c6903205e97c88358ac
---

```yaml
# Model YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/models

type: model
materialize: true
incremental: true
change_mode: patch
timeout: 12h

dev:
  partitions:
    connector: s3
    glob: s3://rill-customer-bucket/dataset/2025/09/02/00/exchange-7f778d56f5-zplb*.parquet

partitions:
  connector: s3
  glob: s3://rill-customer-bucket/dataset/*/*/02/*/exchange*.parquet


connector: clickhouse
sql: |
  SELECT
      toStartOfHour(parseDateTimeBestEffort(timestamp)) AS rolled_up_hour,
      toDateTime(ifNull(parseDateTimeBestEffortOrNull(timestamp), toDateTime(now()))) AS ttl_time,
      now() AS __load_time,
      COUNT(*) AS __sourcerows,

      type,
      app_id,
      app_domain,
      app_name,
      country,
      region,
      zip_code,
      brand_id,
      advertiser_id,
      device_platform_id,
      device_platform,
      adomain,
      iab_cat,
      content_id,
      content_type,
      content_title,

      -- numeric sums (Int*/Float*)
      sumIf(cpm, cpm IS NOT NULL) AS cpm_sum,
      sumIf(bidfloor, bidfloor IS NOT NULL) AS bidfloor_sum,
      sumIf(floor_percentage, floor_percentage IS NOT NULL) AS floor_percentage_sum,
      sumIf(revShare, revShare IS NOT NULL) AS revShare_sum,
      sumIf(rts_num_video, rts_num_video IS NOT NULL) AS rts_num_video_sum,
      sumIf(rts_num_audio, rts_num_audio IS NOT NULL) AS rts_num_audio_sum,
      sumIf(content_len, content_len IS NOT NULL) AS content_len_sum,
  FROM s3(
      '{{.split.uri}}',
      '{{ .env.aws_access_key_id }}',
      '{{ .env.aws_secret_access_key }}',
      'Parquet'
  )
  GROUP BY ALL
  ORDER BY rolled_up_hour


prod:
  output:
    connector: clickhouse
    engine: "ReplicatedMergeTree()"
    order_by: ttl_time
    TTL: "toDateTime(ttl_time) + INTERVAL 1 DAY TO VOLUME 's3_cached',
          toDateTime(ttl_time) + INTERVAL 30 DAY TO VOLUME 's3_direct'
          SETTINGS
            ttl_only_drop_parts=1,
            storage_policy='s3_cached_tiered_policy';"
```
