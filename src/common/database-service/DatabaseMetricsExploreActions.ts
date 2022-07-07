import { DatabaseActions } from "$common/database-service/DatabaseActions";
import type { DatabaseMetadata } from "$common/database-service/DatabaseMetadata";
import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import {
  getExpressionColumnsFromMeasures,
  getFilterFromFilters,
  normaliseMeasures,
} from "./utils";
import type { BasicMeasureDefinition } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";

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
    filters: ActiveValues
  ) {
    // remove filters for this specific dimension.
    const isolatedFilters = { ...filters };
    delete isolatedFilters[column];
    const whereClause =
      filters && Object.keys(isolatedFilters).length
        ? `AND ${getFilterFromFilters(isolatedFilters)}`
        : "";
    return this.databaseClient.execute(`
      SELECT ${expression} as value, "${column}" as label from "${table}"
      WHERE "${column}" IS NOT NULL ${whereClause}
      GROUP BY "${column}"
      ORDER BY value desc
      LIMIT 15
    `);
  }

  public async getBigNumber(
    metadata: DatabaseMetadata,
    table: string,
    measures: Array<BasicMeasureDefinition>,
    filters: ActiveValues
  ): Promise<BigNumberResponse> {
    measures = normaliseMeasures(measures);
    const whereClause =
      filters && Object.keys(filters).length
        ? `WHERE ${getFilterFromFilters(filters)}`
        : "";
    const bigNumbers = await this.databaseClient.execute(
      `
      SELECT ${getExpressionColumnsFromMeasures(measures)} from "${table}"
      ${whereClause}
    `
    );
    return { bigNumbers: bigNumbers?.[0] };
  }
}
