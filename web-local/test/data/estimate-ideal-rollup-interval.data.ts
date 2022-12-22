import type { DataProviderData } from "@adityahegde/typescript-test-utils";
import { PreviewRollupInterval } from "@rilldata/web-common/lib/duckdb-data-types";

export interface EstimatedRollupIntervalTestCase {
  start: string;
  end: string;
  interval: string;
  expectedRollupInterval: string;
}

/**
 * This data definition is used to create synthetic timestamp columns that we
 * then use to estimate a good rollup interval.
 */
export const timeGrainSeriesData: DataProviderData<
  [EstimatedRollupIntervalTestCase]
> = {
  subData: [
    {
      title: "millisecond rollup",
      subData: [
        {
          args: [
            {
              start: "2021-01-01 00:00:00",
              end: "2021-01-01 00:00:01",
              interval: "1 millisecond",
              expectedRollupInterval: PreviewRollupInterval.ms,
            },
          ],
        },
        {
          args: [
            {
              start: "2021-01-01 00:00:00",
              end: "2021-01-01 00:01:00",
              interval: "1 millisecond",
              expectedRollupInterval: PreviewRollupInterval.ms,
            },
          ],
        },
      ],
    },
    {
      title: "second rollup",
      subData: [
        {
          args: [
            {
              start: "2021-01-01 00:00:00",
              end: "2021-01-01 00:01:01",
              interval: "1 millisecond",
              expectedRollupInterval: PreviewRollupInterval.second,
            },
          ],
        },
        {
          args: [
            {
              start: "2021-01-01 00:00:00",
              end: "2021-01-01 01:00:00",
              interval: "1 second",
              expectedRollupInterval: PreviewRollupInterval.second,
            },
          ],
        },
      ],
    },
    {
      title: "minute rollup",
      subData: [
        {
          args: [
            {
              start: "2021-01-01 00:00:00",
              end: "2021-01-01 23:59:59",
              interval: "1 second",
              expectedRollupInterval: PreviewRollupInterval.minute,
            },
          ],
        },
      ],
    },
    {
      title: "hour rollup",
      subData: [
        {
          args: [
            {
              start: "2021-01-01 00:00:00",
              end: "2021-01-07 00:00:00",
              interval: "1 hour",
              expectedRollupInterval: PreviewRollupInterval.hour,
            },
          ],
        },
      ],
    },
    {
      title: "day rollup",
      subData: [
        {
          args: [
            {
              start: "2021-01-01 00:00:00",
              end: "2040-12-27 23:59:59.999", // note: leap days make the difference here!
              interval: "1 hour",
              expectedRollupInterval: PreviewRollupInterval.day,
            },
          ],
        },
      ],
    },
    {
      title: "month rollup",
      subData: [
        {
          args: [
            {
              start: "2021-01-01 00:00:00",
              end: "2520-01-01 00:00:00",
              interval: "day",
              expectedRollupInterval: PreviewRollupInterval.month,
            },
          ],
        },
      ],
    },
    {
      title: "year rollup",
      subData: [
        {
          args: [
            {
              start: "500-01-01 00:00:00",
              end: "2021-01-01 00:00:00",
              interval: "1 year",
              expectedRollupInterval: PreviewRollupInterval.year,
            },
          ],
        },
      ],
    },
  ],
};
