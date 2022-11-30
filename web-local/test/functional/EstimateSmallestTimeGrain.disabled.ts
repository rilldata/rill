import { DataProviderData, TestBase } from "@adityahegde/typescript-test-utils";
import { FunctionalTestBase } from "./FunctionalTestBase";
import type { DatabaseService } from "@rilldata/web-local/common/database-service/DatabaseService";
import type { EstimatedSmallestTimeGrain } from "@rilldata/web-local/common/database-service/DatabaseColumnActions";
import { dataModelerServiceFactory } from "@rilldata/web-local/server/serverFactory";

import { getTestConfig } from "../utils/getTestConfig";
import { generateSeries } from "../utils/query-generators";
import { timeGrainSeriesData } from "../data/EstimatedSmallestTimeGrain.data";
import type { GeneratedTimeseriesTestCase } from "../data/EstimatedSmallestTimeGrain.data";
import type { DuckDBClient } from "@rilldata/web-local/common/database-service/DuckDBClient";

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
