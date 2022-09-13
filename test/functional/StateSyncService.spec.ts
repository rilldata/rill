import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import { DataModelerStateSyncService } from "$common/data-modeler-state-service/sync-service/DataModelerStateSyncService";
import { asyncWait } from "$common/utils/waitUtils";
import { dataModelerServiceFactory } from "$server/serverFactory";
import { TestBase } from "@adityahegde/typescript-test-utils";
import { execSync } from "node:child_process";
import { UserColumnsTestData } from "../data/DataLoader.data";
import {
  SingleTableQuery,
  SingleTableQueryColumnsTestData,
  TwoTableJoinQuery,
  TwoTableJoinQueryColumnsTestData,
} from "../data/ModelQuery.data";
import { getTestConfig } from "../utils/getTestConfig";
import { FunctionalTestBase } from "./FunctionalTestBase";

const SYNC_TEST_FOLDER = "temp/sync-test";

@FunctionalTestBase.Suite
export class StateSyncServiceSpec extends FunctionalTestBase {
  protected secondDataModelerStateService: DataModelerStateService;
  protected secondDataModelerService: DataModelerService;
  protected secondDataModelerSyncService: DataModelerStateSyncService;

  public async setup(): Promise<void> {
    execSync(`rm -rf ${SYNC_TEST_FOLDER}/*`);
    execSync(`mkdir -p ${SYNC_TEST_FOLDER}`);
    const config = getTestConfig(SYNC_TEST_FOLDER, {
      profileWithUpdate: false,
    });
    await super.setup(config);

    const secondServerInstances = dataModelerServiceFactory(config);
    this.secondDataModelerStateService =
      secondServerInstances.dataModelerStateService;
    this.secondDataModelerService = secondServerInstances.dataModelerService;
    this.secondDataModelerSyncService = new DataModelerStateSyncService(
      config,
      this.secondDataModelerStateService.entityStateServices,
      this.secondDataModelerService,
      this.secondDataModelerStateService
    );
    await this.secondDataModelerService.init();
    await this.secondDataModelerSyncService.init();

    await this.secondDataModelerService.dispatch("addOrUpdateTableFromFile", [
      "test/data/AdBids.parquet",
    ]);
    await this.secondDataModelerService.dispatch("addOrUpdateTableFromFile", [
      "test/data/AdImpressions.parquet",
    ]);
    await asyncWait(100);
  }

  @FunctionalTestBase.Test()
  public async clientShouldPickupNewTables(): Promise<void> {
    let instances = this.getTables("name", "Users");
    expect(instances[0]).toBeUndefined();

    await this.secondDataModelerService.dispatch("addOrUpdateTableFromFile", [
      "test/data/Users.parquet",
    ]);
    await asyncWait(500);

    instances = this.getTables("name", "Users");
    expect(instances[0].name).toBe("Users");
    this.assertColumns(instances[1].profile, UserColumnsTestData);
  }

  @FunctionalTestBase.Test()
  public async clientShouldPickupModelUpdates(): Promise<void> {
    await this.secondDataModelerService.dispatch("addModel", [
      { name: "newModel", query: SingleTableQuery },
    ]);
    await asyncWait(1000);
    await this.waitForModels();
    const [model, derivedModel] = this.getModels("tableName", "newModel");
    expect(model.name).toBe("newModel.sql");
    this.assertColumns(derivedModel.profile, SingleTableQueryColumnsTestData);

    await this.clientDataModelerService.dispatch("updateModelQuery", [
      model.id,
      TwoTableJoinQuery,
    ]);
    await asyncWait(1000);
    await this.waitForModels();
    const [, updatedDerivedModel] = this.getModels("tableName", "newModel");
    expect(model.name).toBe("newModel.sql");
    this.assertColumns(
      updatedDerivedModel.profile,
      TwoTableJoinQueryColumnsTestData
    );
  }

  // @FunctionalTestBase.Test()
  // There is no parallel delete just yet. We should fix this in the future
  public async clientShouldPickupModelDeletion(): Promise<void> {
    await this.secondDataModelerService.dispatch("addModel", [
      { name: "newModelDelete", query: SingleTableQuery },
    ]);
    await asyncWait(1000);
    await this.waitForModels();
    const [model] = this.getModels("tableName", "newModelDelete");
    expect(model.name).toBe("newModelDelete.sql");

    await this.clientDataModelerService.dispatch("deleteModel", [model.id]);
    await asyncWait(1000);
    await this.waitForModels();
    const updatedModels = this.getModels("tableName", "newModelDelete");
    expect(updatedModels[0]).toBeUndefined();
    expect(updatedModels[1]).toBeUndefined();
  }

  @TestBase.AfterSuite()
  public async teardown(): Promise<void> {
    await super.teardown();
    await this.secondDataModelerSyncService?.destroy();
    await this.secondDataModelerService?.destroy();
  }
}
