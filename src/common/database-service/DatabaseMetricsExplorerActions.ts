import type { BasicMeasureDefinition } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { DatabaseActions } from "$common/database-service/DatabaseActions";
import type { DatabaseMetadata } from "$common/database-service/DatabaseMetadata";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import type { MetricViewRequestFilter } from "$common/rill-developer-service/MetricViewActions";
import {
  getExpressionColumnsFromMeasures,
  getWhereClauseFromFilters,
  normaliseMeasures,
} from "./utils";

export interface BigNumberResponse {
  id?: string;
  bigNumbers: Record<string, number>;
  error?: string;
}

export class DatabaseMetricsExplorerActions extends DatabaseActions {
  public async getLeaderboardValues(
    metadata: DatabaseMetadata,
    table: string,
    column: string,
    expression: string,
    filters: MetricViewRequestFilter,
    timestampColumn: string,
    timeRange?: TimeSeriesTimeRange
  ) {
    // remove filters for this specific dimension.
    const isolatedFilters = { ...filters };
    delete isolatedFilters[column];

    const whereClause = getWhereClauseFromFilters(
      filters,
      timestampColumn,
      timeRange,
      "WHERE"
    );

    return this.databaseClient.execute(
      `
      SELECT ${expression} as value, "${column}" as label from "${table}"
      ${whereClause}
      GROUP BY "${column}"
      ORDER BY value desc NULLS LAST
      LIMIT 15
    `
    );
  }

  public async getBigNumber(
    metadata: DatabaseMetadata,
    table: string,
    measures: Array<BasicMeasureDefinition>,
    filters: MetricViewRequestFilter,
    timestampColumn: string,
    timeRange?: TimeSeriesTimeRange
  ): Promise<BigNumberResponse> {
    measures = normaliseMeasures(measures);

    const whereClause = getWhereClauseFromFilters(
      filters,
      timestampColumn,
      timeRange,
      "WHERE"
    );

    try {
      const bigNumbers = await this.databaseClient.execute<
        Record<string, number>
      >(
        `
        SELECT ${getExpressionColumnsFromMeasures(measures)} from "${table}"
        ${whereClause}
      `
      );
      return { bigNumbers: bigNumbers?.[0] };
    } catch (err) {
      return {
        bigNumbers: {},
        error: err.message,
      };
    }
  }

  public async validateMeasureExpression(
    metadata: DatabaseMetadata,
    table: string,
    expression: string
  ): Promise<string> {
    try {
      await this.databaseClient.prepare(`select ${expression} from ${table}`);
    } catch (err) {
      return err.message;
    }
    return "";
  }
}
