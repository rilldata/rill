import { FunctionalTestBase } from "./FunctionalTestBase";
import { exec } from "node:child_process";
import { promisify } from "util";
import { existsSync, readFileSync } from "fs";
import {
  AdBidsColumnsTestData,
  AdImpressionColumnsTestData,
} from "../data/DataLoader.data";
import type { DerivedTableState } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import type { PersistentTableState } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { CLI_COMMAND } from "../utils/getCliCommand";

const execPromise = promisify(exec);

const CLI_TEST_FOLDER = "temp/test-cli";
const CLI_STATE_FOLDER = `${CLI_TEST_FOLDER}/state`;
const CLI_TEST_FOLDER_ARG = `--project ${CLI_TEST_FOLDER}`;

@FunctionalTestBase.Suite
export class CLISpec extends FunctionalTestBase {
  @FunctionalTestBase.BeforeSuite()
  public async setup() {
    // override parent method to stop it from starting local server
  }

  @FunctionalTestBase.BeforeEachTest()
  public async setupTests(): Promise<void> {
    await execPromise(`rm -rf ${CLI_TEST_FOLDER}`);
  }

  @FunctionalTestBase.Test()
  public async shouldInitProject(): Promise<void> {
    await execPromise(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);
    expect(existsSync(`${CLI_TEST_FOLDER}/persistent_table_state.json`));
    expect(existsSync(`${CLI_TEST_FOLDER}/stage.db`));
  }

  @FunctionalTestBase.Test()
  public async shouldAddTables(): Promise<void> {
    await execPromise(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);
    await execPromise(
      `${CLI_COMMAND} import-source test/data/AdBids.csv ${CLI_TEST_FOLDER_ARG}`
    );
    await execPromise(
      `${CLI_COMMAND} import-source test/data/AdImpressions.tsv --name Impressions ${CLI_TEST_FOLDER_ARG}`
    );

    const persistentState: PersistentTableState = JSON.parse(
      readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString()
    );
    const derivedState: DerivedTableState = JSON.parse(
      readFileSync(`${CLI_STATE_FOLDER}/derived_table_state.json`).toString()
    );
    expect(persistentState.entities[0].name).toBe("AdBids");
    this.assertColumns(derivedState.entities[0].profile, AdBidsColumnsTestData);
    expect(persistentState.entities[1].name).toBe("Impressions");
    this.assertColumns(
      derivedState.entities[1].profile,
      AdImpressionColumnsTestData
    );
  }

  @FunctionalTestBase.Test()
  public async shouldErrorIfSourceFileIsMalformed(): Promise<void> {
    await execPromise(`${CLI_COMMAND} init ${CLI_TEST_FOLDER}`);
    await execPromise(
      `${CLI_COMMAND} import-source test/data/AdBids.parquet ${CLI_TEST_FOLDER_ARG}`
    );
    // import the broken dataset.
    await execPromise(
      `${CLI_COMMAND} import-source test/data/BrokenCSV.csv ${CLI_TEST_FOLDER_ARG}`
    );

    let persistentState: PersistentTableState = JSON.parse(
      readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString()
    );

    // BrokenCSV should not be present in the state.
    const brokenCSVState = persistentState.entities.find(
      (entity) => entity.tableName === "BrokenCSV"
    );
    expect(brokenCSVState).toBeUndefined();

    // let's get the state for AdBids before we attempt to import a broken dataset into it.
    const adBids = persistentState.entities.find(
      (entity) => entity.tableName === "AdBids"
    );
    // let's try to replace AdBids
    await execPromise(
      `${CLI_COMMAND} import-source test/data/BrokenCSV.csv --name AdBids --force ${CLI_TEST_FOLDER_ARG}`
    );
    // check to see if the sources are the same.
    persistentState = JSON.parse(
      readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString()
    );

    const newAdBids = persistentState.entities.find(
      (entity) => entity.tableName === "AdBids"
    );
    const oldStateObject = { ...adBids };

    const oldStateUpdateTime = oldStateObject.lastUpdated;
    delete oldStateObject.lastUpdated;
    // check the newAdBids field minus the lastUpdated time stamp.
    expect(newAdBids).toEqual(expect.objectContaining(oldStateObject));
    expect(newAdBids.lastUpdated > oldStateUpdateTime).toBeTruthy();
  }

  @FunctionalTestBase.Test()
  public async shouldDropTable(): Promise<void> {
    await execPromise(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);
    await execPromise(
      `${CLI_COMMAND} import-source test/data/AdBids.parquet ${CLI_TEST_FOLDER_ARG}`
    );

    let persistentState: PersistentTableState = JSON.parse(
      readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString()
    );
    expect(persistentState.entities[0].name).toBe("AdBids");

    await execPromise(
      `${CLI_COMMAND} drop-source AdBids ${CLI_TEST_FOLDER_ARG}`
    );

    persistentState = JSON.parse(
      readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString()
    );
    expect(persistentState.entities.length).toBe(0);
  }
}
