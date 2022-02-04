import {DuckDBAPI} from "./DuckDBAPI";
import type {CategoricalSummary, NumericHistogramBin, NumericSummary, TimeRangeSummary} from "$lib/types";

const TOP_K_COUNT = 50;

export class DuckDBColumnAPI extends DuckDBAPI {
    public async getTopKAndCardinality(tableName: string, columnName: string,
                                       func = "count(*)"): Promise<CategoricalSummary> {
        return {
            topK: await this.getTopK(tableName, columnName, func),
            cardinality: await this.getCardinality(tableName, columnName),
        };
    }

    public async getNullCount(tableName: string, columnName: string): Promise<number> {
        const [nullity] = await this.duckDBClient.all(
            `SELECT COUNT(*) as count FROM ${tableName} WHERE ${columnName} IS NULL;`);
        return nullity.count;
    }

    public async getDescriptiveStatistics(tableName: string, columnName: string): Promise<NumericSummary> {
        const [results] = await this.duckDBClient.all(`
            SELECT
                min(${columnName}) as min,
                reservoir_quantile(${columnName}, 0.25) as q25,
                reservoir_quantile(${columnName}, 0.5)  as q50,
                reservoir_quantile(${columnName}, 0.75) as q75,
                max(${columnName}) as max,
                avg(${columnName})::FLOAT as mean,
                stddev_pop(${columnName}) as sd
            FROM ${tableName};
       `);
        return { statistics: results };
    }

    public async getNumericHistogram(tableName: string, field: string, fieldType: string): Promise<NumericSummary> {
        const buckets = await this.duckDBClient.all(`SELECT count(*) as count, ${field} FROM ${tableName} WHERE ${field} IS NOT NULL GROUP BY ${field} USING SAMPLE reservoir(1000 ROWS);`)
        const bucketSize = Math.min(40, buckets.length);
        const results = await this.duckDBClient.all(`
            WITH dataset AS (
                SELECT ${fieldType === 'TIMESTAMP' ? `epoch(${field})` : `${field}::DOUBLE`} as ${field} FROM ${tableName}
            ) , S AS (
                SELECT 
                    min(${field}) as minVal,
                    max(${field}) as maxVal,
                    (max(${field}) - min(${field})) as range
                    FROM dataset
            ), values AS (
                SELECT ${field} as value from dataset
                WHERE ${field} IS NOT NULL
            ), buckets AS (
                SELECT
                    range as bucket,
                    (range) * (select range FROM S) / ${bucketSize} + (select minVal from S) as low,
                    (range + 1) * (select range FROM S) / ${bucketSize} + (select minVal from S) as high
                FROM range(0, ${bucketSize}, 1)
            )
            , histogram_stage AS (
                SELECT
                    bucket,
                    low,
                    high,
                    count(values.value) as count
                FROM buckets
                LEFT JOIN values ON values.value BETWEEN low and high
                GROUP BY bucket, low, high
                ORDER BY BUCKET
            )
            SELECT 
                bucket,
                low,
                high,
                CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + 1 ELSE count END AS count
                FROM histogram_stage;
        `);
        return { histogram: results };
    }

    public async getTimeRange(tableName: string, columnName: string): Promise<TimeRangeSummary> {
        const [ranges] = await this.duckDBClient.all(`
	        SELECT
		    min(${columnName}) as min, max(${columnName}) as max, 
		    max(${columnName}) - min(${columnName}) as interval
		    FROM ${tableName};
	    `)
        return ranges;
    }

    private async getTopK(tableName: string, columnName: string, func = "count(*)"): Promise<any> {
        return this.duckDBClient.all(`
            SELECT ${columnName} as value, ${func} AS count from ${tableName}
            GROUP BY ${columnName}
            ORDER BY count desc
            LIMIT ${TOP_K_COUNT};
        `);
    }

    private async getCardinality(tableName: string, columnName: string): Promise<number> {
        const [results] = await this.duckDBClient.all(
            `SELECT approx_count_distinct(${columnName}) as count from ${tableName};`);
        return results.count;
    }
}
