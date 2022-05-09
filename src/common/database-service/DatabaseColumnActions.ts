import {DatabaseActions} from "./DatabaseActions";
import type {CategoricalSummary, NumericSummary, TimeRangeSummary} from "$lib/types";
import type {DatabaseMetadata} from "$common/database-service/DatabaseMetadata";
import {sanitizeColumn} from "$common/utils/queryUtils";
import {TIMESTAMPS} from "$lib/duckdb-data-types";

const TOP_K_COUNT = 50;

export enum TimeGrain {
  milliseconds = "milliseconds",
  seconds = "seconds",
  minutes = "minutes",
  hours = "hours",
  days = "days",
  weeks = "weeks",
  months = "months",
  years = "years"
}

const MICROS = {
  hour:   1000 * 1000 * 60 * 60,
  minute: 1000 * 1000 * 60,
  second: 1000 * 1000,
  millisecond: 1000
}

/**
 * All database column actions return javascript objects that get folded 
 * into a `summary` field in the derived table. Thus any action in this file must
 * return an object.
 */
export class DatabaseColumnActions extends DatabaseActions {
    public async getTopKAndCardinality(metadata: DatabaseMetadata, tableName: string, columnName: string,
                                       func = "count(*)"): Promise<CategoricalSummary> {
        return {
            topK: await this.getTopKOfColumn(metadata, tableName, columnName, func),
            cardinality: await this.getCardinalityOfColumn(metadata, tableName, columnName),
        };
    }

    public async getNullCount(metadata: DatabaseMetadata,
                              tableName: string, columnName: string): Promise<number> {
        const sanitizedColumName = sanitizeColumn(columnName);
        const [nullity] = await this.databaseClient.execute(
            `SELECT COUNT(*) as count FROM '${tableName}' WHERE ${sanitizedColumName} IS NULL;`);
        return nullity.count;
    }

    public async getDescriptiveStatistics(metadata: DatabaseMetadata,
                                          tableName: string, columnName: string): Promise<NumericSummary> {
        const sanitizedColumnName = sanitizeColumn(columnName);
        const [results] = await this.databaseClient.execute(`
            SELECT
                min(${sanitizedColumnName}) as min,
                reservoir_quantile(${sanitizedColumnName}, 0.25) as q25,
                reservoir_quantile(${sanitizedColumnName}, 0.5)  as q50,
                reservoir_quantile(${sanitizedColumnName}, 0.75) as q75,
                max(${sanitizedColumnName}) as max,
                avg(${sanitizedColumnName})::FLOAT as mean,
                stddev_pop(${sanitizedColumnName}) as sd
            FROM '${tableName}';
       `);
        return { statistics: results };
    }

    /**
     * Estimates a reasonable rollup timegrain for the given table & timestamp column.
     * This is currently based on a heuristic method that largely looks at
     * the time range of the timestamp column and guesses a good rollup grain.
     */
    public async estimateRollupTimegrain(metadata: DatabaseMetadata,
        tableName: string, columnName: string): Promise<any> {
        
        function rollupTimegrainReturnFormat(rollupGranularity, minValue, maxValue) {
          return {
            rollupGranularity, minValue, maxValue
          }
        }

        const [timeRange] =  await this.databaseClient.execute(`SELECT 
            max("${columnName}") - min("${columnName}") as r,
            max("${columnName}") as max_value,
            min("${columnName}") as min_value,
            count(*) as count
            from 
        ${tableName}`);

        const { r, max_value: maxValue, min_value: minValue, count } = timeRange;
        
        let range = typeof r === 'number' ? {days: r, micros:0, months: 0} : r;
    
        if (range.days === 0 && range.micros <= MICROS.minute) {
            return rollupTimegrainReturnFormat('1 millisecond', minValue, maxValue);
        }
    
        if (range.days === 0 && range.micros > MICROS.minute && range.micros <= MICROS.minute * 60) {
            return rollupTimegrainReturnFormat('1 second', minValue, maxValue);
        }
    
        if (range.days === 0 && range.micros <= MICROS.hour * 24) {
            return rollupTimegrainReturnFormat('1 minute', minValue, maxValue);
        }
        if (range.days < 7) {
            return rollupTimegrainReturnFormat('1 hour', minValue, maxValue);
        }
        if (range.days < 365) {
            return rollupTimegrainReturnFormat('1 day', minValue, maxValue);
        }
    
        if (range.days < (365 * 20) && count > range.days * 15) {
            return rollupTimegrainReturnFormat('1 day', minValue, maxValue);
        }
        if (range.days < (365 * 20)) {
            return rollupTimegrainReturnFormat('1 day', minValue, maxValue);
        }
        if (range.days < (365 * 500)) {
            return rollupTimegrainReturnFormat('1 month', minValue, maxValue);
        } 
        return rollupTimegrainReturnFormat('1 year', minValue, maxValue);
    }

    /**
     * A single-pass heuristic for generating a count(*) over an entire timestamp column,
     * rolled up to a hopefully useful timegrain.
     * It will make a reasonable estimate of how the column should be rolled up,
     * then produce both the final rolled up result and a reduced M4-like spark representation
     * using the same temporary table.
     * A sampleSize argument is provided to provide an "optimistic query" option for the user
     * if speed is a concern. A reasonable 1,000,000 row sample should speed things up
     * in extreme cases.
     */
    public async estimateTimestampRollup(
          metadata: DatabaseMetadata,
          table:string, column:string, pixels = undefined, sampleSize = undefined) {
        const {rollupGranularity, minValue, maxValue} = await this.estimateRollupTimegrain(metadata, table, column);
        const [ totalRow ] = await this.databaseClient.execute(`SELECT count(*) as c from "${table}"`);
        const total = totalRow.c;
        
        const inflator = (sampleSize && sampleSize < total) ? (total / sampleSize) : 1;
        /**
         * Generate the rolled up time series as a temporary table and 
         * then compute the result set + any M4-like reduction on it.
         * We first create a resultset of zero-values,
         * then join this result set against the empirical counts.
         */
        try {
            await this.databaseClient.execute(`CREATE TEMPORARY TABLE _ts_ AS (
                WITH template as (
                    SELECT 
                        generate_series as ts 
                    FROM 
                        generate_series(
                            date_trunc(
                                '${rollupGranularity.split(' ')[1]}', 
                                TIMESTAMP '${minValue.toISOString()}'
                            ), 
                            date_trunc(
                                '${rollupGranularity.split(' ')[1]}', 
                                TIMESTAMP '${maxValue.toISOString()}'
                            ), 
                            interval ${rollupGranularity})
                ),
                transformed AS (
                    SELECT 
                        date_trunc('${rollupGranularity.split(' ')[1]}', "${column}") as ts 
                    FROM "${table}"
                        ${sampleSize && sampleSize < total ? `USING SAMPLE ${(sampleSize / total) * 100}%` : ''}
                ),
                series AS (
                    SELECT count(*) as count, ts from transformed 
                    GROUP BY ts ORDER BY ts
                )
                SELECT COALESCE(series.count * ${inflator}::FLOAT, 0) as count, template.ts from template
                LEFT OUTER JOIN series ON template.ts = series.ts
                ORDER BY template.ts
            )`);
        } catch (err) {
            await this.databaseClient.execute(`DROP TABLE IF EXISTS _ts_;`);
        }
        
        // decide if the final result set has to be thrown out
        const [{ count }] = await this.databaseClient.execute(`
            SELECT 
                count(*) as count
                from _ts_`);
        
        let results;
        let spark;
        
        if (pixels) {
            /**
             * Generate the M4-like reduction of this time series.
             * This variation will produce 4 points per pixel – the left bound, right bound,
             * the max, and the min.
             */
            spark = await this.databaseClient.execute(`
            WITH Q as (
                SELECT extract('epoch' from ts) as t, "count" as v from _ts_
            ),
            M as (
                SELECT min(t) as t1, max(t) as t2, max(t) - min(t) as diff FROM Q
            )
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
        
                round(${pixels} * (t - (select t1 from M)) / (select diff from M)) AS bin
        
            FROM Q GROUP BY bin
            ORDER BY bin
            `)
            spark = spark.map((di => {
                /** 
                 * Extract the four prototype points for each pixel bin,
                 * sort the points, then flatten the entire array.
                 */
                let points = [
                    {
                        ts: new Date(di.min_t),
                        count: di.argmin_tv, bin: di.bin
                    },
                    {
                        ts: new Date(di.argmin_vt),
                        count: di.min_v, bin: di.bin
                    },
                    {
                        ts: new Date(di.argmax_vt),
                        count: di.max_v , bin: di.bin
                    },
                    {
                        ts: new Date(di.max_t),
                        count: di.argmax_tv, bin: di.bin
                    },
                ]
                /** Sort the final point set. */
                points = points.sort((a,b) => {
                    if (a.ts === b.ts) return 0;
                    return a.ts < b.ts ? -1 : 1;
                })
                return points;
            })).flat();
        } 
        /** Materialize the final time series. */
        results = await this.databaseClient.execute(`SELECT * from _ts_`);
        await this.databaseClient.execute(`DROP TABLE _ts_`);

        return {
            rollup: {
              results, 
              granularity: rollupGranularity, 
              reduced: (pixels && pixels * 4 <= count),
              rows: total,
              spark,
              sampleSize: sampleSize
            }
        }
    }

    /**
     * Estimates the smallest time grain present in the column.
     * The "smallest time grain" is the smallest value that we believe the user
     * can reliably roll up. In other words, if the data is reported daily, this
     * action will return "day", since that's the smallest rollup grain we can
     * rely on.
     * 
     * This function can only focus on some common time grains. It will operate on
     * - ms
     * - second
     * - minute
     * - hour
     * - day
     * - week
     * - month
     * - year
     * 
     * It will not estimate any more nuanced or difficult-to-measure time grains, such as
     * quarters, once-a-month, etc.
     * 
     * It accomplishes its goal by sampling 500k values of a column and then estimating the cardinality
     * of each. If there are < 500k samples, the action will use all of the column's data.
     * We're not sure all the ways this heuristic will fail, but it seems pretty resilient to the tests
     * we've thrown at it.
     */
    public async estimateSmallestTimeGrain(metadata: DatabaseMetadata,
        tableName: string, columnName: string, sampleSize = 500000): Promise<{ estimatedSmallestTimeGrain: TimeGrain }> {
      const [total] = await this.databaseClient.execute(`
        SELECT count(*) as c from "${tableName}"
      `)
      const totalRows = total.c;
      // only sample when you have a lot of data.
      const useSample = sampleSize > totalRows ? '' : `USING SAMPLE ${(100 * sampleSize / totalRows)}%`

      const [ timeGrainResult ] = await this.databaseClient.execute(`
      WITH cleaned_column AS (
          SELECT "${columnName}" as cd
          from ${tableName}
          ${useSample}
      ),
      time_grains as (
      SELECT 
          approx_count_distinct(extract('years' from cd)) as year,
          approx_count_distinct(extract('months' from cd)) as month,
          approx_count_distinct(extract('dayofyear' from cd)) as dayofyear,
          approx_count_distinct(extract('dayofmonth' from cd)) as dayofmonth,
          min(cd = last_day(cd)) = TRUE as lastdayofmonth,
          approx_count_distinct(extract('weekofyear' from cd)) as weekofyear,
          approx_count_distinct(extract('dayofweek' from cd)) as dayofweek,
          approx_count_distinct(extract('hour' from cd)) as hour,
          approx_count_distinct(extract('minute' from cd)) as minute,
          approx_count_distinct(extract('second' from cd)) as second,
          approx_count_distinct(extract('millisecond' from cd) - extract('seconds' from cd) * 1000) as ms
      FROM cleaned_column
      )
      SELECT 
        COALESCE(
            case WHEN ms > 1 THEN 'milliseconds' else NULL END,
            CASE WHEN second > 1 THEN 'seconds' else NULL END,
            CASE WHEN minute > 1 THEN 'minutes' else null END,
            CASE WHEN hour > 1 THEN 'hours' else null END,
            -- cases above, if equal to 1, then we have some candidates for
            -- bigger time grains. We need to reverse from here
            -- years, months, weeks, days.
            CASE WHEN dayofyear = 1 and year > 1 THEN 'years' else null END,
            CASE WHEN (dayofmonth = 1 OR lastdayofmonth) and month > 1 THEN 'months' else null END,
            CASE WHEN dayofweek = 1 and weekofyear > 1 THEN 'weeks' else null END,
            CASE WHEN hour = 1 THEN 'days' else null END
        ) as estimatedSmallestTimeGrain
      FROM time_grains
      `);
      return timeGrainResult;
    }

    public async getNumericHistogram(metadata: DatabaseMetadata,
                                              tableName: string, columnName: string, columnType: string): Promise<NumericSummary> {
        const sanitizedColumnName = sanitizeColumn(columnName);
        // use approx_count_distinct to get the immediate cardinality of this column.
        const [buckets] = await this.databaseClient.execute(`SELECT approx_count_distinct(${sanitizedColumnName}) as count from ${tableName}`);
        const bucketSize = Math.min(40, buckets.count);
        const result = await this.databaseClient.execute(`
          WITH data_table AS (
            SELECT ${TIMESTAMPS.has(columnType) ? `epoch(${sanitizedColumnName})` : `${sanitizedColumnName}::DOUBLE`} as ${sanitizedColumnName} 
            FROM ${tableName}
            WHERE ${sanitizedColumnName} IS NOT NULL
          ), S AS (
            SELECT 
              min(${sanitizedColumnName}) as minVal,
              max(${sanitizedColumnName}) as maxVal,
              (max(${sanitizedColumnName}) - min(${sanitizedColumnName})) as range
              FROM data_table
          ), values AS (
            SELECT ${sanitizedColumnName} as value from data_table
            WHERE ${sanitizedColumnName} IS NOT NULL
          ), buckets AS (
            SELECT
              range as bucket,
              (range) * (select range FROM S) / ${bucketSize} + (select minVal from S) as low,
              (range + 1) * (select range FROM S) / ${bucketSize} + (select minVal from S) as high
            FROM range(0, ${bucketSize}, 1)
          ),
          histogram_stage AS (
          SELECT
              bucket,
              low,
              high,
              count(values.value) as count
            FROM buckets
            LEFT JOIN values ON (values.value >= low and values.value < high)
            GROUP BY bucket, low, high
            ORDER BY BUCKET
          ),
          -- calculate the right edge, sine in histogram_stage we don't look at the values that
          -- might be the largest.
          right_edge AS (
            SELECT count(*) as c from values WHERE value = (select maxVal from S)
          )
          SELECT 
            bucket,
            low,
            high,
            -- fill in the case where we've filtered out the highest value and need to recompute it, otherwise use count.
            CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + (select c from right_edge) ELSE count END AS count
            FROM histogram_stage
	      `);
        return { histogram: result };
    }

    public async getTimeRange(metadata: DatabaseMetadata,
                              tableName: string, columnName: string): Promise<TimeRangeSummary> {
        const sanitizedColumnName = sanitizeColumn(columnName);
        const [ranges] = await this.databaseClient.execute(`
	        SELECT
		    min(${sanitizedColumnName}) as min, max(${sanitizedColumnName}) as max, 
		    max(${sanitizedColumnName}) - min(${sanitizedColumnName}) as interval
		    FROM '${tableName}';
	    `);
        return ranges;
    }

    private async getTopKOfColumn(metadata: DatabaseMetadata,
                          tableName: string, columnName: string, func = "count(*)"): Promise<any> {
        const sanitizedColumnName = sanitizeColumn(columnName);
        return this.databaseClient.execute(`
            SELECT ${sanitizedColumnName} as value, ${func} AS count from '${tableName}'
            GROUP BY ${sanitizedColumnName}
            ORDER BY count desc
            LIMIT ${TOP_K_COUNT};
        `);
    }

    private async getCardinalityOfColumn(metadata: DatabaseMetadata,
                                 tableName: string, columnName: string): Promise<number> {
        const sanitizedColumnName = sanitizeColumn(columnName);
        const [results] = await this.databaseClient.execute(
            `SELECT approx_count_distinct(${sanitizedColumnName}) as count from '${tableName}';`);
        return results.count;
    }
}
