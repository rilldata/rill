import {DatabaseActions} from "./DatabaseActions";
import type {CategoricalSummary, NumericSummary, TimeRangeSummary} from "$lib/types";
import {calculateBins} from "$common/utils/calculateBins";

const TOP_K_COUNT = 50;

export class DatabaseColumnActions extends DatabaseActions {
    public async getTopKAndCardinality(tableName: string, columnName: string,
                                       func = "count(*)"): Promise<CategoricalSummary> {
        return {
            topK: await this.getTopK(tableName, columnName, func),
            cardinality: await this.getCardinality(tableName, columnName),
        };
    }

    public async getNullCount(tableName: string, columnName: string): Promise<number> {
        const [nullity] = await this.dbClient.execute(
            `SELECT COUNT(*) as count FROM '${tableName}' WHERE ${columnName} IS NULL;`);
        return nullity.count;
    }

    public async getDescriptiveStatistics(tableName: string, columnName: string): Promise<NumericSummary> {
        const [results] = await this.dbClient.execute(`
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

    public async getNumericHistogram(tableName: string, field: string, fieldType: string): Promise<NumericSummary> {
        const results = await this.dbClient.execute(
            `SELECT ${fieldType === 'TIMESTAMP' ? `epoch(${field})` : `${field}::DOUBLE`} as ${field} FROM '${tableName}'`);
        return { histogram: calculateBins(results, field) };
    }

    public async getTimeRange(tableName: string, columnName: string): Promise<TimeRangeSummary> {
        const [ranges] = await this.dbClient.execute(`
	        SELECT
		    min(${columnName}) as min, max(${columnName}) as max, 
		    max(${columnName}) - min(${columnName}) as interval
		    FROM '${tableName}';
	    `);
        return ranges;
    }

    private async getTopK(tableName: string, columnName: string, func = "count(*)"): Promise<any> {
        return this.dbClient.execute(`
            SELECT ${columnName} as value, ${func} AS count from '${tableName}'
            GROUP BY ${columnName}
            ORDER BY count desc
            LIMIT ${TOP_K_COUNT};
        `);
    }

    private async getCardinality(tableName: string, columnName: string): Promise<number> {
        const [results] = await this.dbClient.execute(
            `SELECT approx_count_distinct(${columnName}) as count from '${tableName}';`);
        return results.count;
    }
}
