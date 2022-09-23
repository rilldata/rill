import { DataProviderData, TestBase } from "@adityahegde/typescript-test-utils";
import { FunctionalTestBase } from "./FunctionalTestBase";
import type { DatabaseService } from "$web-local/common/database-service/DatabaseService";
import type { EstimatedSmallestTimeGrain } from "$web-local/common/database-service/DatabaseColumnActions";
import { RootConfig } from "$web-local/common/config/RootConfig";
import { DatabaseConfig } from "$web-local/common/config/DatabaseConfig";
import { StateConfig } from "$web-local/common/config/StateConfig";
import { dataModelerServiceFactory } from "$web-local/server/serverFactory";

import { generateSeries } from "../utils/query-generators";
import { timeGrainSeriesData } from "../data/EstimatedSmallestTimeGrain.data";
import type { GeneratedTimeseriesTestCase } from "../data/EstimatedSmallestTimeGrain.data";
import type { DuckDBClient } from "$web-local/common/database-service/DuckDBClient";

const SYNC_TEST_FOLDER = "temp/sync-test";

@FunctionalTestBase.Suite
export class EstimateSmallestTimeGrainSpec extends FunctionalTestBase {
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
    [GeneratedTimeseriesTestCase]
  > {
    return timeGrainSeriesData;
  }

  @TestBase.Test("seriesGeneratedTimegrainData")
  public async shouldIdentifyTimegrain(args: GeneratedTimeseriesTestCase) {
    await this.dbClient.execute(
      generateSeries(args.table, args.start, args.end, args.interval)
    );
    const result = (await this.databaseService.dispatch(
      "estimateSmallestTimeGrain",
      [args.table, "ts"]
    )) as { estimatedSmallestTimeGrain: EstimatedSmallestTimeGrain };
    expect(args.expectedTimeGrain).toBe(result.estimatedSmallestTimeGrain);
  }
}
