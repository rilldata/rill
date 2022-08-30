import { DatabaseConfig } from "$common/config/DatabaseConfig";
import { RootConfig } from "$common/config/RootConfig";
import { StateConfig } from "$common/config/StateConfig";
import { expect } from "@playwright/test";
import { existsSync, readFileSync, writeFileSync } from "fs";
import { execSync } from "node:child_process";
import {
  NestedQuery,
  NestedQueryColumnsTestData,
  SingleTableQuery,
  SingleTableQueryColumnsTestData,
  TwoTableJoinQuery,
  TwoTableJoinQueryColumnsTestData,
} from "../data/ModelQuery.data";
import { FunctionalTestBase } from "./FunctionalTestBase";
import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";

const SYNC_TEST_FOLDER = "temp/model-sync-test";
const MODEL_FOLDER = `${SYNC_TEST_FOLDER}/models`;
const MODEL_0_FILE = `${MODEL_FOLDER}/model_0.sql`;
const MODEL_1_FILE = `${MODEL_FOLDER}/model_1.sql`;

@FunctionalTestBase.Suite
export class ModelFileSyncSpec extends FunctionalTestBase {
  public async setup() {
    execSync(`rm -rf ${SYNC_TEST_FOLDER}`);
    await super.setup(
      new RootConfig({
        database: new DatabaseConfig({ databaseName: ":memory:" }),
        state: new StateConfig({ autoSync: true, syncInterval: 50 }),
        projectFolder: SYNC_TEST_FOLDER,
        profileWithUpdate: true,
      })
    );
  }

  @FunctionalTestBase.BeforeSuite()
  public async setupTables(): Promise<void> {
    await this.loadTestTables();
  }

  @FunctionalTestBase.BeforeEachTest()
  public async setupTests(): Promise<void> {
    await this.clientDataModelerService.dispatch("clearAllModels", []);
  }

  @FunctionalTestBase.Test()
  public async shouldUpdateModelFileAndViceVersa() {
    expect(existsSync(MODEL_0_FILE)).toBe(false);

    await this.clientDataModelerService.dispatch("addModel", [
      { name: "model_0", query: "" },
    ]);
    await this.waitForModels();
    expect(existsSync(MODEL_0_FILE)).toBe(true);

    const [model] = this.getModels("tableName", "model_0");
    await this.clientDataModelerService.dispatch("updateModelQuery", [
      model.id,
      SingleTableQuery,
    ]);
    await this.waitForModels();

    // updating query from client should update the file
    expect(readFileSync(MODEL_0_FILE).toString()).toBe(SingleTableQuery);
    const [, persistentModel] = this.getModels("tableName", "model_0");
    this.assertColumns(
      persistentModel.profile,
      SingleTableQueryColumnsTestData
    );

    // updating query from file should update profiling data
    writeFileSync(MODEL_0_FILE, TwoTableJoinQuery);
    await this.waitForModels();
    const [, newPersistentModel] = this.getModels("tableName", "model_0");
    this.assertColumns(
      newPersistentModel.profile,
      TwoTableJoinQueryColumnsTestData
    );
  }

  @FunctionalTestBase.Test()
  public async shouldDeleteModelOnFileDelete() {
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "model_0", query: SingleTableQuery },
    ]);
    await this.waitForModels();
    expect(existsSync(MODEL_0_FILE)).toBe(true);

    // file is recreated if deleted.
    execSync(`rm ${MODEL_0_FILE}`);
    await this.waitForModels();
    const [model] = this.getModels("tableName", "model_0");
    expect(model).toBe(undefined);
    expect(existsSync(MODEL_0_FILE)).toBe(false);
  }

  @FunctionalTestBase.Test()
  public async shouldRenameModelOnFileCaseRename() {
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "model_0", query: SingleTableQuery },
    ]);
    await this.waitForModels();
    expect(existsSync(MODEL_0_FILE)).toBe(true);
    expect(existsSync(MODEL_1_FILE)).toBe(false);

    // file is renamed if deleted.
    execSync(`mv ${MODEL_0_FILE} ${MODEL_1_FILE}`);
    await this.waitForModels();
    const [model0] = this.getModels("tableName", "model_0");
    const [model1, persistentModel1] = this.getModels("tableName", "model_1");
    expect(model0).toBe(undefined);
    expect(existsSync(MODEL_0_FILE)).toBe(false);
    expect(model1.query).toBe(SingleTableQuery);
    this.assertColumns(
      persistentModel1.profile,
      SingleTableQueryColumnsTestData
    );
    expect(existsSync(MODEL_1_FILE)).toBe(true);
  }

  @FunctionalTestBase.Test()
  public async shouldAddNewModelsOnModelRename() {
    const MODEL_00_FILE = `${MODEL_FOLDER}/model_00.sql`;
    const MODEL_10_FILE = `${MODEL_FOLDER}/model_10.sql`;

    await this.clientDataModelerService.dispatch("addModel", [
      { name: "model_0", query: SingleTableQuery },
    ]);
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "model_1", query: TwoTableJoinQuery },
    ]);
    await this.waitForModels();
    expect(existsSync(MODEL_0_FILE)).toBe(true);
    expect(existsSync(MODEL_00_FILE)).toBe(false);
    expect(existsSync(MODEL_1_FILE)).toBe(true);
    expect(existsSync(MODEL_10_FILE)).toBe(false);

    const [model0] = this.getModels("tableName", "model_0");
    const [model1] = this.getModels("tableName", "model_1");

    // rename model_0 => model_00, model_1 => model_10 then add a new file model_0.sql
    await this.clientDataModelerService.dispatch("updateModelName", [
      model0.id,
      "model_00",
    ]);
    await this.clientDataModelerService.dispatch("updateModelName", [
      model1.id,
      "model_10",
    ]);
    await this.waitForModels();
    writeFileSync(MODEL_0_FILE, NestedQuery);
    await this.waitForModels();
    expect(readFileSync(MODEL_0_FILE).toString()).toBe(NestedQuery);
    expect(readFileSync(MODEL_00_FILE).toString()).toBe(SingleTableQuery);
    expect(existsSync(MODEL_1_FILE)).toBe(false);
    expect(readFileSync(MODEL_10_FILE).toString()).toBe(TwoTableJoinQuery);

    const [, persistentModel0] = this.getModels("tableName", "model_00");
    const [, persistentModel1] = this.getModels("tableName", "model_10");
    const [, persistentModel2] = this.getModels("tableName", "model_0");
    this.assertColumns(
      persistentModel0?.profile,
      SingleTableQueryColumnsTestData
    );
    this.assertColumns(
      persistentModel1?.profile,
      TwoTableJoinQueryColumnsTestData
    );
    this.assertColumns(persistentModel2?.profile, NestedQueryColumnsTestData);
  }

  @FunctionalTestBase.Test()
  public async shouldDeleteNonSqlFiles() {
    const INVALID_FILE = "model_0.sq";
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "model_0", query: SingleTableQuery },
    ]);
    await this.waitForModels();
    expect(existsSync(MODEL_0_FILE)).toBe(true);

    // file is renamed to invalid file. model is deleted
    execSync(`mv ${MODEL_0_FILE} ${INVALID_FILE}`);
    await this.waitForModels();
    let [model] = this.getModels("tableName", "model_0");
    expect(model).toBe(undefined);
    expect(existsSync(MODEL_0_FILE)).toBe(false);
    // invalid file is not deleted
    expect(existsSync(INVALID_FILE)).toBe(true);

    // file is renamed back to .sql file
    execSync(`mv ${INVALID_FILE} ${MODEL_0_FILE}`);
    await this.waitForModels();
    [model] = this.getModels("tableName", "model_0");
    expect(model.tableName).toBe("model_0");
    expect(readFileSync(MODEL_0_FILE).toString()).toBe(SingleTableQuery);
    expect(existsSync(INVALID_FILE)).toBe(false);
  }

  @FunctionalTestBase.Test()
  public async shouldSuccessfullyRenameUpdatingCase() {
    const MODEL_UPPER_0_FILE = `${MODEL_FOLDER}/MODEL_0.sql`;

    await this.clientDataModelerService.dispatch("addModel", [
      { name: "model_0", query: SingleTableQuery },
    ]);
    await this.waitForModels();
    const [model] = this.getModels("tableName", "model_0");

    await this.clientDataModelerService.dispatch("updateModelName", [
      model.id,
      "MODEL_0",
    ]);
    await this.waitForModels();
    // this will be true on a case insensitive file system
    // expect(existsSync(MODEL_0_FILE)).toBe(false);
    expect(existsSync(MODEL_UPPER_0_FILE)).toBe(true);
    expect(readFileSync(MODEL_UPPER_0_FILE).toString()).toBe(SingleTableQuery);
  }

  @FunctionalTestBase.Test()
  public async shouldNotRenameOtherModelWithSameCase() {
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "model_0", query: SingleTableQuery },
    ]);
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "model_1", query: TwoTableJoinQuery },
    ]);
    await this.waitForModels();

    const [model] = this.getModels("tableName", "model_1");

    // try to rename to another model but with a different case
    const resp = await this.clientDataModelerService.dispatch(
      "updateModelName",
      [model.id, "MODEL_0"]
    );
    expect(resp.status).toBe(ActionStatus.Failure);
    await this.waitForModels();

    expect(existsSync(MODEL_0_FILE)).toBe(true);
    expect(readFileSync(MODEL_0_FILE).toString()).toBe(SingleTableQuery);
    expect(existsSync(MODEL_1_FILE)).toBe(true);
    expect(readFileSync(MODEL_1_FILE).toString()).toBe(TwoTableJoinQuery);
  }
}
