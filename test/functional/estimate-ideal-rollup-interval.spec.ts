import { DataProviderData, TestBase } from "@adityahegde/typescript-test-utils";
import { FunctionalTestBase } from "./FunctionalTestBase";
import type { DatabaseService } from "$common/database-service/DatabaseService";
import { RootConfig } from "$common/config/RootConfig";
import { DatabaseConfig } from "$common/config/DatabaseConfig";
import { StateConfig } from "$common/config/StateConfig";
import { dataModelerServiceFactory } from "$server/serverFactory";
import type { DuckDBClient } from "$common/database-service/DuckDBClient";

import { generateSeries } from "../utils/query-generators";
import { timeGrainSeriesData } from "../data/estimate-ideal-rollup-interval.data";
import type { EstimatedRollupIntervalTestCase } from "../data/estimate-ideal-rollup-interval.data";

const SYNC_TEST_FOLDER = "temp/sync-test";

/**
 * NOTE: this test suite may end up getting moved to src/common/dtabase-service/tests.
 * I'll keep it here for this PR until review time / when we decide where we want tests to be.
 */
@FunctionalTestBase.Suite
export class EstimateIdealRollupInterval extends FunctionalTestBase {
  protected databaseService: DatabaseService;
  protected dbClient: DuckDBClient;

  public async setup(): Promise<void> {
    const config = new RootConfig({
      database: new DatabaseConfig({ databaseName: ":memory:" }),
      state: new StateConfig({ autoSync: true, syncInterval: 50 }),
      projectFolder: SYNC_TEST_FOLDER,
      profileWithUpdate: false,
    });
    const secondServerInstances = dataModelerServiceFactory(config);
    this.databaseService =
      secondServerInstances.dataModelerService.getDatabaseService();
    await this.databaseService.init();
    this.dbClient = this.databaseService.getDatabaseClient();
  }

  public seriesGeneratedTimegrainData(): DataProviderData<
    [EstimatedRollupIntervalTestCase]
  > {
    return timeGrainSeriesData;
  }

  @TestBase.Test("seriesGeneratedTimegrainData")
  public async shouldEstimateInterval(args: EstimatedRollupIntervalTestCase) {
    /** create a _test view with a single ts column */
    await this.dbClient.execute(
      generateSeries("_test", args.start, args.end, args.interval)
    );
    /** roll up our _test.ts column */
    const result = await this.databaseService.dispatch(
      "estimateIdealRollupInterval",
      ["_test", "ts"]
    );
    /** drop the temporarily-made view */
    await this.dbClient.execute(`DROP VIEW _test`);
    expect(args.expectedRollupInterval).toBe(result.rollupInterval);
  }
}
