import type { BasicMeasureDefinition } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { DatabaseActions } from "$common/database-service/DatabaseActions";
import type { RollupInterval } from "$common/database-service/DatabaseColumnActions";
import { MICROS } from "$common/database-service/DatabaseColumnActions";
import type { DatabaseMetadata } from "$common/database-service/DatabaseMetadata";
import {
  getCoalesceStatementsMeasures,
  getExpressionColumnsFromMeasures,
  getFilterFromFilters,
  normaliseMeasures,
} from "$common/database-service/utils";
import { PreviewRollupInterval } from "$lib/duckdb-data-types";
import type { ActiveValues } from "$lib/application-state-stores/explorer-stores";

export type TimeSeriesValue = {
  ts: string;
  bin?: number;
} & Record<string, number>;

export interface TimeSeriesResponse {
  id?: string;
  results: Array<TimeSeriesValue>;
  spark?: Array<TimeSeriesValue>;
  timeRange?: TimeSeriesTimeRange;
  sampleSize?: number;
  error?: string;
}
export interface TimeSeriesRollup {
  rollup: TimeSeriesResponse;
}

export enum TimeRangeName {
  LastHour = "Last hour",
  Last6Hours = "Last 6 hours",
  LastDay = "Last day",
  Last2Days = "Last 2 days",
  Last5Days = "Last 5 days",
  LastWeek = "Last week",
  Last2Weeks = "Last 2 weeks",
  Last30Days = "Last 30 days",
  Last60Days = "Last 60 days",
  AllTime = "All time",
  // Today = "Today",
  // MonthToDate = "Month to date",
  // CustomRange = "Custom range",
}

export const lastXTimeRanges: TimeRangeName[] = [
  TimeRangeName.LastHour,
  TimeRangeName.Last6Hours,
  TimeRangeName.LastDay,
  TimeRangeName.Last2Days,
  TimeRangeName.Last5Days,
  TimeRangeName.LastWeek,
  TimeRangeName.Last2Weeks,
  TimeRangeName.Last30Days,
  TimeRangeName.Last60Days,
];

// The string values must adhere to DuckDB INTERVAL syntax, since, in some places, we interpolate an SQL queries with these values.
export enum TimeGrain {
  OneMinute = "1 minute",
  // FiveMinutes = "5 minute",
  // FifteenMinutes = "15 minute",
  OneHour = "1 hour",
  OneDay = "1 day",
  OneWeek = "7 day",
  OneMonth = "1 month",
  OneYear = "1 year",
}
export interface TimeSeriesTimeRange {
  name?: TimeRangeName;
  start?: string;
  end?: string;
  interval?: string; // TODO: switch this to TimeGrain
}

interface TimeseriesReductionQueryResponse {
  bin: number;
  min_t: number;
  argmin_tv: number;
  min_v: number;
  argmin_vt: number;
  max_v: number;
  argmax_vt: number;
  max_t: number;
  argmax_tv: number;
}

export class DatabaseTimeSeriesActions extends DatabaseActions {
  /**
   * A single-pass heuristic for generating a `expression` or count(*) over an entire timestamp column,
   * rolled up to a hopefully useful timegrain.
   * It will make a reasonable estimate of how the column should be rolled up,
   * then produce both the final rolled up result and a reduced M4-like spark representation
   * using the same temporary table.
   * A sampleSize argument is provided to provide an "optimistic query" option for the user
   * if speed is a concern. A reasonable 1,000,000 row sample should speed things up
   * in extreme cases.
   */
  public async generateTimeSeries(
    metadata: DatabaseMetadata,
    {
      tableName,
      measures,
      timestampColumn,
      timeRange,
      filters,
      pixels,
      sampleSize,
    }: {
      tableName: string;
      measures?: Array<BasicMeasureDefinition>;
      timestampColumn: string;
      timeRange?: TimeSeriesTimeRange;
      filters?: ActiveValues;
      pixels?: number;
      sampleSize?: number;
    }
  ): Promise<TimeSeriesRollup> {
    measures = normaliseMeasures(measures);

    timeRange = await this.getNormalisedTimeRange(
      tableName,
      timestampColumn,
      timeRange
    );

    let timeGranularity = timeRange.interval.split(" ")[1];
    // add workaround for weekly. DuckDB does not support
    // a 1 week syntax, so in the case that we have 7 day, let's use
    // the week timeGranularity for truncation.
    if (timeRange.interval === "7 day") {
      timeGranularity = "week";
    }

    const filter =
      filters && Object.keys(filters).length > 0
        ? " WHERE " + getFilterFromFilters(filters)
        : "";

    /**
     * Generate the rolled up time series as a temporary table and
     * then compute the result set + any M4-like reduction on it.
     * We first create a resultset of zero-values,
     * then join this result set against the empirical counts.
     *
     * Limitation: due to the use of `date_trunc()` in the `series` CTE,
     * this query cannot handle a DuckDB interval that uses
     * n>1 unit, e.g. 15 minutes, 7 days, etc. See this StackOverflow answer
     * for a different approach: https://stackoverflow.com/a/41944083
     */
    try {
      await this.databaseClient.execute(
        `CREATE TEMPORARY TABLE _ts_ AS (
        -- generate a time series column that has the intended range
        WITH template as (
          SELECT 
            generate_series as ts 
          FROM 
            generate_series(
              date_trunc(
                '${timeGranularity}', 
                TIMESTAMP '${timeRange.start}'
              ), 
              date_trunc(
                '${timeGranularity}', 
                TIMESTAMP '${timeRange.end}'
              ),
              interval ${timeRange.interval})
        ),
        -- transform the original data, and optionally sample it.
        series AS (
          SELECT 
            date_trunc('${timeGranularity}', "${timestampColumn}") as ts,
            ${getExpressionColumnsFromMeasures(measures)}
          FROM "${tableName}" ${filter}
          GROUP BY ts ORDER BY ts
        )
        -- join the transformed data with the generated time series column,
        -- coalescing the first value to get the 0-default when the rolled up data
        -- does not have that value.
        SELECT 
          ${getCoalesceStatementsMeasures(measures)},
          template.ts from template
        LEFT OUTER JOIN series ON template.ts = series.ts
        ORDER BY template.ts
      )`
      );
    } catch (err) {
      console.error(err);
      await this.databaseClient.execute(`DROP TABLE IF EXISTS _ts_;`);
      return {
        rollup: {
          results: [],
          timeRange,
          ...(pixels ? { spark: [] } : {}),
          sampleSize,
          error: err.message,
        },
      };
    }

    let spark;

    if (pixels) {
      /**
       * Generate the M4-like reduction of this time series.
       * This variation will produce 4 points per pixel – the left bound, right bound,
       * the max, and the min.
       */
      spark = await this.createTimestampRollupReduction(
        metadata,
        "_ts_",
        "ts",
        "count",
        pixels
      );
    }

    const results = await this.databaseClient.execute<TimeSeriesValue>(
      `SELECT * from _ts_`
    );
    await this.databaseClient.execute(`DROP TABLE _ts_`);

    return {
      rollup: {
        results,
        timeRange,
        spark,
        sampleSize,
      },
    };
  }

  /**
   * Contains an as-of-this-commit unpublished algorithm for an M4-like line density reduction.
   * This will take in an n-length time series and produce a pixels * 4 reduction of the time series
   * that preserves the shape and trends.
   *
   * This algorithm expects the source table to have a timestamp column and some kind of value column,
   * meaning it expects the data to essentially already be aggregated.
   *
   * It's important to note that this implemention is NOT the original M4 aggregation method, but a method
   * that has the same basic understanding but is much faster.
   *
   * Nonetheless, we mostly use this to reduce a many-thousands-point-long time series to about 120 * 4 pixels.
   * Importantly, this function runs very fast. For more information about the original M4 method,
   * see http://www.vldb.org/pvldb/vol7/p797-jugel.pdf
   */
  public async createTimestampRollupReduction(
    metadata: DatabaseMetadata,
    table: string,
    timestampColumn: string,
    valueColumn: string,
    pixels: number
  ) {
    const [timeSeriesLength] = await this.databaseClient.execute(`
        SELECT count(*) as c FROM "${table}"
    `);
    if (timeSeriesLength.c < pixels * 4) {
      return this.databaseClient.execute(`
          SELECT "${timestampColumn}" as ts, "${valueColumn}" as count FROM "${table}"
      `);
    }

    const reduction = await this.databaseClient
      .execute<TimeseriesReductionQueryResponse>(`
      -- extract unix time
      WITH Q as (
        SELECT extract('epoch' from "${timestampColumn}") as t, "${valueColumn}" as v FROM "${table}"
      ),
      -- generate bounds
      M as (
        SELECT min(t) as t1, max(t) as t2, max(t) - min(t) as diff FROM Q
      )
      -- core logic
      SELECT 
        -- left boundary point
        min(t) * 1000  as min_t, 
        arg_min(v, t) as argmin_tv, 

        -- right boundary point
        max(t) * 1000 as max_t, 
        arg_max(v, t) as argmax_tv,

        -- smallest point within boundary
        min(v) as min_v, 
        arg_min(t, v) * 1000  as argmin_vt,

        -- largest point within boundary
        max(v) as max_v, 
        arg_max(t, v) * 1000  as argmax_vt,

        round(${pixels} * (t - (SELECT t1 FROM M)) / (SELECT diff FROM M)) AS bin
  
      FROM Q GROUP BY bin
      ORDER BY bin
    `);

    return reduction
      .map((di) => {
        /**
         * Extract the four prototype points for each pixel bin,
         * sort the points, then flatten the entire array.
         */
        let points = [
          {
            ts: new Date(di.min_t),
            count: di.argmin_tv,
            bin: di.bin,
          },
          {
            ts: new Date(di.argmin_vt),
            count: di.min_v,
            bin: di.bin,
          },
          {
            ts: new Date(di.argmax_vt),
            count: di.max_v,
            bin: di.bin,
          },
          {
            ts: new Date(di.max_t),
            count: di.argmax_tv,
            bin: di.bin,
          },
        ];
        /** Sort the final point set. */
        points = points.sort((a, b) => {
          if (a.ts === b.ts) return 0;
          return a.ts < b.ts ? -1 : 1;
        });
        return points;
      })
      .flat();
  }

  /**
   * Estimates a reasonable rollup timegrain for the given table & timestamp column.
   * This is currently based on a heuristic method that largely looks at
   * the time range of the timestamp column and guesses a good rollup grain.
   * @returns {RollupInterval} the rollup interval information, an object with a rollupInterval, minValue, maxValue
   */
  public async estimateIdealRollupInterval(
    metadata: DatabaseMetadata,
    tableName: string,
    columnName: string
  ): Promise<RollupInterval> {
    function rollupTimegrainReturnFormat(
      rollupInterval,
      minValue,
      maxValue
    ): RollupInterval {
      return {
        rollupInterval,
        minValue,
        maxValue,
      };
    }

    const [timeRange] = await this.databaseClient.execute<{
      r: number;
      max_value: number;
      min_value: number;
      count: number;
    }>(`SELECT 
        max("${columnName}") - min("${columnName}") as r,
        max("${columnName}") as max_value,
        min("${columnName}") as min_value,
        count(*) as count
        from 
      ${tableName}`);

    const { r, max_value: maxValue, min_value: minValue } = timeRange;

    const range = typeof r === "number" ? { days: r, micros: 0, months: 0 } : r;

    if (range.days === 0 && range.micros <= MICROS.minute) {
      return rollupTimegrainReturnFormat(
        PreviewRollupInterval.ms,
        minValue,
        maxValue
      );
    }

    if (
      range.days === 0 &&
      range.micros > MICROS.minute &&
      range.micros <= MICROS.hour
    ) {
      return rollupTimegrainReturnFormat(
        PreviewRollupInterval.second,
        minValue,
        maxValue
      );
    }

    if (range.days === 0 && range.micros <= MICROS.day) {
      return rollupTimegrainReturnFormat(
        PreviewRollupInterval.minute,
        minValue,
        maxValue
      );
    }
    if (range.days <= 7) {
      return rollupTimegrainReturnFormat(
        PreviewRollupInterval.hour,
        minValue,
        maxValue
      );
    }
    if (range.days <= 365 * 20) {
      return rollupTimegrainReturnFormat(
        PreviewRollupInterval.day,
        minValue,
        maxValue
      );
    }
    if (range.days <= 365 * 500) {
      return rollupTimegrainReturnFormat(
        PreviewRollupInterval.month,
        minValue,
        maxValue
      );
    }
    return rollupTimegrainReturnFormat(
      PreviewRollupInterval.year,
      minValue,
      maxValue
    );
  }

  private async getNormalisedTimeRange(
    tableName: string,
    timestampColumn: string,
    timeRange: TimeSeriesTimeRange
  ): Promise<TimeSeriesTimeRange> {
    let rollupInterval = timeRange?.interval;
    if (!rollupInterval) {
      /** NOTE: we will need to put together a different interval function for the explore later. this one will note be flexible enough for user needs */
      const estimatedRollupInterval = await this.estimateIdealRollupInterval(
        undefined,
        tableName,
        timestampColumn
      );
      rollupInterval = estimatedRollupInterval.rollupInterval;
    }

    const [actualTimeRange] = await this.databaseClient.execute<{
      min: number;
      max: number;
    }>(`SELECT
		    min("${timestampColumn}") as min, max("${timestampColumn}") as max 
		    FROM ${tableName}`);

    let startTime = new Date(timeRange?.start || actualTimeRange.min);
    if (Number.isNaN(startTime.getTime())) {
      startTime = new Date(actualTimeRange.min);
    }
    let endTime = new Date(timeRange?.end || actualTimeRange.max);
    if (Number.isNaN(endTime.getTime())) {
      endTime = new Date(actualTimeRange.max);
    }

    return {
      interval: rollupInterval,
      start: startTime.toISOString(),
      end: endTime.toISOString(),
    };
  }
}
