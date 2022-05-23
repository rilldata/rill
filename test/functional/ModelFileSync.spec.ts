import { FunctionalTestBase } from "./FunctionalTestBase";
import { RootConfig } from "$common/config/RootConfig";
import { DatabaseConfig } from "$common/config/DatabaseConfig";
import { StateConfig } from "$common/config/StateConfig";
import {
  NestedQuery,
  NestedQueryColumnsTestData,
  SingleTableQuery,
  SingleTableQueryColumnsTestData,
  TwoTableJoinQuery,
  TwoTableJoinQueryColumnsTestData,
} from "../data/ModelQuery.data";
import { existsSync, readFileSync, writeFileSync } from "fs";
import { expect } from "@playwright/test";
import { execSync } from "node:child_process";

const SYNC_TEST_FOLDER = "temp/model-sync-test";
const MODEL_FOLDER = `${SYNC_TEST_FOLDER}/models`;
const QUERY_0_FILE = `${MODEL_FOLDER}/query_0.sql`;
const QUERY_1_FILE = `${MODEL_FOLDER}/query_1.sql`;

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
    expect(existsSync(QUERY_0_FILE)).toBe(false);

    await this.clientDataModelerService.dispatch("addModel", [
      { name: "query_0", query: "" },
    ]);
    await this.waitForModels();
    expect(existsSync(QUERY_0_FILE)).toBe(true);

    const [model] = this.getModels("tableName", "query_0");
    await this.clientDataModelerService.dispatch("updateModelQuery", [
      model.id,
      SingleTableQuery,
    ]);
    await this.waitForModels();

    // updating query from client should update the file
    expect(readFileSync(QUERY_0_FILE).toString()).toBe(SingleTableQuery);
    const [, persistentModel] = this.getModels("tableName", "query_0");
    this.assertColumns(
      persistentModel.profile,
      SingleTableQueryColumnsTestData
    );

    // updating query from file should update profiling data
    writeFileSync(QUERY_0_FILE, TwoTableJoinQuery);
    await this.waitForModels();
    const [, newPersistentModel] = this.getModels("tableName", "query_0");
    this.assertColumns(
      newPersistentModel.profile,
      TwoTableJoinQueryColumnsTestData
    );
  }

  @FunctionalTestBase.Test()
  public async shouldDeleteModelOnFileDelete() {
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "query_0", query: SingleTableQuery },
    ]);
    await this.waitForModels();
    expect(existsSync(QUERY_0_FILE)).toBe(true);

    // file is recreated if deleted.
    execSync(`rm ${QUERY_0_FILE}`);
    await this.waitForModels();
    const [model] = this.getModels("tableName", "query_0");
    expect(model).toBe(undefined);
    expect(existsSync(QUERY_0_FILE)).toBe(false);
  }

  @FunctionalTestBase.Test()
  public async shouldRenameModelOnFileRename() {
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "query_0", query: SingleTableQuery },
    ]);
    await this.waitForModels();
    expect(existsSync(QUERY_0_FILE)).toBe(true);
    expect(existsSync(QUERY_1_FILE)).toBe(false);

    // file is renamed if deleted.
    execSync(`mv ${QUERY_0_FILE} ${QUERY_1_FILE}`);
    await this.waitForModels();
    const [model0] = this.getModels("tableName", "query_0");
    const [model1, persistentModel1] = this.getModels("tableName", "query_1");
    expect(model0).toBe(undefined);
    expect(existsSync(QUERY_0_FILE)).toBe(false);
    expect(model1.query).toBe(SingleTableQuery);
    this.assertColumns(
      persistentModel1.profile,
      SingleTableQueryColumnsTestData
    );
    expect(existsSync(QUERY_1_FILE)).toBe(true);
  }

  @FunctionalTestBase.Test()
  public async shouldAddNewModelsOnModelRename() {
    const QUERY_00_FILE = `${MODEL_FOLDER}/query_00.sql`;
    const QUERY_10_FILE = `${MODEL_FOLDER}/query_10.sql`;

    await this.clientDataModelerService.dispatch("addModel", [
      { name: "query_0", query: SingleTableQuery },
    ]);
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "query_1", query: TwoTableJoinQuery },
    ]);
    await this.waitForModels();
    expect(existsSync(QUERY_0_FILE)).toBe(true);
    expect(existsSync(QUERY_00_FILE)).toBe(false);
    expect(existsSync(QUERY_1_FILE)).toBe(true);
    expect(existsSync(QUERY_10_FILE)).toBe(false);

    const [model0] = this.getModels("tableName", "query_0");
    const [model1] = this.getModels("tableName", "query_1");

    // rename query_0 => query_00, query_1 => query_10 then add a new file query_0.sql
    await this.clientDataModelerService.dispatch("updateModelName", [
      model0.id,
      "query_00",
    ]);
    await this.clientDataModelerService.dispatch("updateModelName", [
      model1.id,
      "query_10",
    ]);
    await this.waitForModels();
    writeFileSync(QUERY_0_FILE, NestedQuery);
    await this.waitForModels();
    expect(readFileSync(QUERY_0_FILE).toString()).toBe(NestedQuery);
    expect(readFileSync(QUERY_00_FILE).toString()).toBe(SingleTableQuery);
    expect(existsSync(QUERY_1_FILE)).toBe(false);
    expect(readFileSync(QUERY_10_FILE).toString()).toBe(TwoTableJoinQuery);

    const [, persistentModel0] = this.getModels("tableName", "query_00");
    const [, persistentModel1] = this.getModels("tableName", "query_10");
    const [, persistentModel2] = this.getModels("tableName", "query_0");
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
    const INVALID_FILE = "query_0.sq";
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "query_0", query: SingleTableQuery },
    ]);
    await this.waitForModels();
    expect(existsSync(QUERY_0_FILE)).toBe(true);

    // file is renamed to invalid file. model is deleted
    execSync(`mv ${QUERY_0_FILE} ${INVALID_FILE}`);
    await this.waitForModels();
    let [model] = this.getModels("tableName", "query_0");
    expect(model).toBe(undefined);
    expect(existsSync(QUERY_0_FILE)).toBe(false);
    // invalid file is not deleted
    expect(existsSync(INVALID_FILE)).toBe(true);

    // file is renamed back to .sql file
    execSync(`mv ${INVALID_FILE} ${QUERY_0_FILE}`);
    await this.waitForModels();
    [model] = this.getModels("tableName", "query_0");
    expect(model.tableName).toBe("query_0");
    expect(readFileSync(QUERY_0_FILE).toString()).toBe(SingleTableQuery);
    expect(existsSync(INVALID_FILE)).toBe(false);
  }
}
