import type { EstimatedSmallestTimeGrain } from "$common/database-service/DatabaseColumnActions";
import type { DatabaseService } from "$common/database-service/DatabaseService";
import type { DuckDBClient } from "$common/database-service/DuckDBClient";
import { dataModelerServiceFactory } from "$server/serverFactory";
import { DataProviderData, TestBase } from "@adityahegde/typescript-test-utils";
import type { GeneratedTimeseriesTestCase } from "../data/EstimatedSmallestTimeGrain.data";
import { timeGrainSeriesData } from "../data/EstimatedSmallestTimeGrain.data";
import { getTestConfig } from "../utils/getTestConfig";

import { generateSeries } from "../utils/query-generators";
import { FunctionalTestBase } from "./FunctionalTestBase";

const SYNC_TEST_FOLDER = "temp/sync-test";

@FunctionalTestBase.Suite
export class EstimateSmallestTimeGrainSpec extends FunctionalTestBase {
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
