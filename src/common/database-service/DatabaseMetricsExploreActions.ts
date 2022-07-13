import { DatabaseActions } from "$common/database-service/DatabaseActions";
import type { DatabaseMetadata } from "$common/database-service/DatabaseMetadata";
import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import {
  getExpressionColumnsFromMeasures,
  getWhereClauseFromFilters,
  normaliseMeasures,
} from "./utils";
import type { BasicMeasureDefinition } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";

export interface BigNumberResponse {
  id?: string;
  bigNumbers: Record<string, number>;
}

export class DatabaseMetricsExploreActions extends DatabaseActions {
  public async getLeaderboardValues(
    metadata: DatabaseMetadata,
    table: string,
    column: string,
    expression: string,
    filters: ActiveValues,
    timestampColumn: string,
    timeRange?: TimeSeriesTimeRange
  ) {
    // remove filters for this specific dimension.
    const isolatedFilters = { ...filters };
    delete isolatedFilters[column];

    const whereClause = getWhereClauseFromFilters(
      isolatedFilters,
      timestampColumn,
      timeRange,
      "AND"
    );

    return this.databaseClient.execute(
      `
      SELECT ${expression} as value, "${column}" as label from "${table}"
      WHERE "${column}" IS NOT NULL ${whereClause}
      GROUP BY "${column}"
      ORDER BY value desc
      LIMIT 15
    `
    );
  }

  public async getBigNumber(
    metadata: DatabaseMetadata,
    table: string,
    measures: Array<BasicMeasureDefinition>,
    filters: ActiveValues,
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

    const bigNumbers = await this.databaseClient.execute(
      `
      SELECT ${getExpressionColumnsFromMeasures(measures)} from "${table}"
      ${whereClause}
    `
    );
    return { bigNumbers: bigNumbers?.[0] };
  }
}
