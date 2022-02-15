import {DatabaseActions} from "./DatabaseActions";
import type {CategoricalSummary, NumericSummary, TimeRangeSummary} from "$lib/types";
import type {DatabaseMetadata} from "$common/database-service/DatabaseMetadata";

const TOP_K_COUNT = 50;

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
        const [nullity] = await this.databaseClient.execute(
            `SELECT COUNT(*) as count FROM '${tableName}' WHERE ${columnName} IS NULL;`);
        return nullity.count;
    }

    public async getDescriptiveStatistics(metadata: DatabaseMetadata,
                                          tableName: string, columnName: string): Promise<NumericSummary> {
        const [results] = await this.databaseClient.execute(`
            SELECT
                min(${columnName}) as min,
                reservoir_quantile(${columnName}, 0.25) as q25,
                reservoir_quantile(${columnName}, 0.5)  as q50,
                reservoir_quantile(${columnName}, 0.75) as q75,
                max(${columnName}) as max,
                avg(${columnName})::FLOAT as mean,
                stddev_pop(${columnName}) as sd
            FROM '${tableName}';
       `);
        return { statistics: results };
    }

    public async getNumericHistogram(metadata: DatabaseMetadata,
                                              tableName: string, field: string, fieldType: string): Promise<NumericSummary> {
        const buckets = await this.databaseClient.execute(`SELECT count(*) as count, ${field} FROM ${tableName} WHERE ${field} IS NOT NULL GROUP BY ${field} USING SAMPLE reservoir(1000 ROWS);`)
        const bucketSize = Math.min(40, buckets.length);
        const result = await this.databaseClient.execute(`
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
        return { histogram: result };
    }

    public async getTimeRange(metadata: DatabaseMetadata,
                              tableName: string, columnName: string): Promise<TimeRangeSummary> {
        const [ranges] = await this.databaseClient.execute(`
	        SELECT
		    min(${columnName}) as min, max(${columnName}) as max, 
		    max(${columnName}) - min(${columnName}) as interval
		    FROM '${tableName}';
	    `);
        return ranges;
    }

    private async getTopKOfColumn(metadata: DatabaseMetadata,
                          tableName: string, columnName: string, func = "count(*)"): Promise<any> {
        return this.databaseClient.execute(`
            SELECT ${columnName} as value, ${func} AS count from '${tableName}'
            GROUP BY ${columnName}
            ORDER BY count desc
            LIMIT ${TOP_K_COUNT};
        `);
    }

    private async getCardinalityOfColumn(metadata: DatabaseMetadata,
                                 tableName: string, columnName: string): Promise<number> {
        const [results] = await this.databaseClient.execute(
            `SELECT approx_count_distinct(${columnName}) as count from '${tableName}';`);
        return results.count;
    }
}
