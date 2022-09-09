import type { BasicMeasureDefinition } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { DatabaseActions } from "$common/database-service/DatabaseActions";
import type { DatabaseMetadata } from "$common/database-service/DatabaseMetadata";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import type {
  MetricsViewRequestFilter,
  MetricsViewTopListSortEntry,
} from "$common/rill-developer-service/MetricsViewActions";
import {
  getCoalesceExpressionForMeasures,
  getWhereClauseFromFilters,
  normaliseMeasures,
} from "./utils";

export interface BigNumberResponse {
  id?: string;
  bigNumbers: Record<string, number>;
  error?: string;
}

export interface LeaderboardQueryAdditionalArguments {
  filters: MetricsViewRequestFilter;
  sort: Array<MetricsViewTopListSortEntry>;
  timestampColumn: string;
  timeRange: TimeSeriesTimeRange;
  limit: number;
}

export class DatabaseMetricsExplorerActions extends DatabaseActions {
  public async getLeaderboardValues(
    metadata: DatabaseMetadata,
    table: string,
    column: string,
    measures: Array<BasicMeasureDefinition>,
    // additional arguments
    {
      filters,
      sort,
      timestampColumn,
      timeRange,
      limit,
    }: LeaderboardQueryAdditionalArguments
  ) {
    limit ??= 15;

    // remove filters for this specific dimension.
    const isolatedFilters: MetricsViewRequestFilter = {
      include: filters?.include.filter((filter) => filter.name !== column),
      exclude: filters?.exclude.filter((filter) => filter.name !== column),
    };

    const whereClause = getWhereClauseFromFilters(
      isolatedFilters,
      timestampColumn,
      timeRange,
      "WHERE"
    );

    const expressionColumns = measures
      .map((measure) => `${measure.expression} as ${measure.sqlName}`)
      .join(",");
    const sortQuery = sort?.length
      ? "ORDER BY " +
        sort
          .map(
            (sortEntry) => `${sortEntry.name} ${sortEntry.direction} NULLS LAST`
          )
          .join(",")
      : "";

    return this.databaseClient.execute(
      `
      SELECT ${expressionColumns}, "${column}" from "${table}"
      ${whereClause}
      GROUP BY "${column}"
      ${sortQuery}
      LIMIT ${limit}
    `
    );
  }

  public async getBigNumber(
    metadata: DatabaseMetadata,
    table: string,
    measures: Array<BasicMeasureDefinition>,
    filters: MetricsViewRequestFilter,
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
        SELECT ${getCoalesceExpressionForMeasures(measures)} from "${table}"
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
