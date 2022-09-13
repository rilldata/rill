import type { DatabaseService } from "$common/database-service/DatabaseService";
import type { DuckDBClient } from "$common/database-service/DuckDBClient";
import { dataModelerServiceFactory } from "$server/serverFactory";
import { DataProviderData, TestBase } from "@adityahegde/typescript-test-utils";
import type { EstimatedRollupIntervalTestCase } from "../data/estimate-ideal-rollup-interval.data";
import { timeGrainSeriesData } from "../data/estimate-ideal-rollup-interval.data";
import { getTestConfig } from "../utils/getTestConfig";

import { generateSeries } from "../utils/query-generators";
import { FunctionalTestBase } from "./FunctionalTestBase";

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
    const config = getTestConfig(SYNC_TEST_FOLDER, {
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
    const result = (await this.databaseService.dispatch(
      "estimateIdealRollupInterval",
      ["_test", "ts"]
    )) as { rollupInterval: string };
    /** drop the temporarily-made view */
    await this.dbClient.execute(`DROP VIEW _test`);
    expect(args.expectedRollupInterval).toBe(result.rollupInterval);
  }
}
