import { DatabaseActions } from "$common/database-service/DatabaseActions";
import type { DatabaseMetadata } from "$common/database-service/DatabaseMetadata";
import type { ActiveValues } from "$lib/redux-store/metrics-leaderboard-slice";

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
        ? `AND ${this.getFilterFromFilters(isolatedFilters)}`
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
    expression: string,
    filters: ActiveValues
  ) {
    const whereClause =
      filters && Object.keys(filters).length
        ? `WHERE ${this.getFilterFromFilters(filters)}`
        : "";
    return this.databaseClient.execute(`
      SELECT ${expression} as value from "${table}"
      ${whereClause};
    `);
  }

  private getFilterFromFilters(filters: ActiveValues): string {
    return Object.keys(filters)
      .map((field) => {
        return filters[field]
          .map(([value, filterType]) =>
            filterType ? `"${field}" = '${value}'` : `"${field}" != '${value}'`
          )
          .join(" OR ");
      })
      .join(" AND ");
  }
}
