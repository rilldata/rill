import { DataProviderData, TestBase } from "@adityahegde/typescript-test-utils";
import { FunctionalTestBase } from "./FunctionalTestBase";
import type { DatabaseService } from "@rilldata/web-local/common/database-service/DatabaseService";
import { RootConfig } from "@rilldata/web-local/common/config/RootConfig";
import { DatabaseConfig } from "@rilldata/web-local/common/config/DatabaseConfig";
import { StateConfig } from "@rilldata/web-local/common/config/StateConfig";
import { dataModelerServiceFactory } from "@rilldata/web-local/server/serverFactory";
import type { DuckDBClient } from "@rilldata/web-local/common/database-service/DuckDBClient";

import { generateSeries } from "../utils/query-generators";

const SYNC_TEST_FOLDER = "temp/sync-test";

export interface RollupReductionTestCase {
  start: string;
  end: string;
  interval: string;
  pixels: number;
  notEnoughPoints?: boolean;
}

/**
 * This data definition is used to create synthetic timestamp columns that we
 * then use to estimate a good rollup interval.
 */
export const rollupReduction: DataProviderData<[RollupReductionTestCase]> = {
  subData: [
    {
      title: "should have more source points than pixels * 4",
      subData: [
        {
          args: [
            {
              start: "2021-01-01 00:00:00",
              end: "2021-01-01 00:09:59",
              interval: "1 millisecond",
              pixels: 100,
            },
          ],
        },
      ],
    },
    {
      title: "shoudl have fewer source points than pixels * 4",
      subData: [
        {
          args: [
            {
              start: "2021-01-01 00:00:00",
              end: "2021-01-01 00:00:59",
              interval: "1 second",
              pixels: 100,
              notEnoughPoints: true,
            },
          ],
        },
      ],
    },
  ],
};

/**
 * NOTE: this test suite may end up getting moved to src/common/database-service/tests.
 * I'll keep it here for this PR until review time / when we decide where we want tests to be.
 */
@FunctionalTestBase.Suite
export class CreateTimestampRollupReduction extends FunctionalTestBase {
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

  public rollupReductionTestCase(): DataProviderData<
    [RollupReductionTestCase]
  > {
    return rollupReduction;
  }

  @TestBase.Test("rollupReductionTestCase")
  public async shouldEstimateInterval(args: RollupReductionTestCase) {
    /** create a _test view with a single ts column */
    await this.dbClient.execute(
      generateSeries("_test", args.start, args.end, args.interval, true)
    );

    /** roll up our _test.ts column */
    const result = (await this.databaseService.dispatch(
      "createTimestampRollupReduction",
      ["_test", "ts", "count", args.pixels]
    )) as Array<unknown>;

    /** drop the temporarily-made view */
    await this.dbClient.execute(`DROP VIEW _test`);
    if (args.notEnoughPoints) {
      expect((args.pixels + 1) * 4).toBeGreaterThan(result.length);
    } else {
      expect((args.pixels + 1) * 4).toBe(result.length);
    }
  }
}
