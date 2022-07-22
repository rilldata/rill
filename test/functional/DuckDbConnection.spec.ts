import type { DerivedTableState } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import type { PersistentTableState } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { exec } from "node:child_process";
import { readFileSync } from "fs";
import {
  AdBidsColumnsTestData,
  AdImpressionColumnsTestData,
  UserColumnsTestData,
} from "../data/DataLoader.data";
import { promisify } from "util";
import { FunctionalTestBase } from "./FunctionalTestBase";
import { CLI_COMMAND } from "../utils/getCliCommand";

const execPromise = promisify(exec);

const CLI_TEST_FOLDER = "temp/test-duckdb-import";
const CLI_STATE_FOLDER = `${CLI_TEST_FOLDER}/state`;
const CLI_TEST_FOLDER_ARG = `--project ${CLI_TEST_FOLDER}`;

const CLI_TEST_DUCKDB_FOLDER = "temp/test-duckdb";
const CLI_STATE_DUCKDB_FOLDER = `${CLI_TEST_DUCKDB_FOLDER}/state`;
const CLI_TEST_DUCKDB_FILE = `${CLI_TEST_DUCKDB_FOLDER}/stage.db`;
const CLI_TEST_DUCKDB_FOLDER_ARG = `--project ${CLI_TEST_DUCKDB_FOLDER}`;

@FunctionalTestBase.Suite
export class DuckDbConnectionSpec extends FunctionalTestBase {
  @FunctionalTestBase.BeforeSuite()
  public async setup() {
    // override parent method to stop it from starting local server
  }

  @FunctionalTestBase.BeforeEachTest()
  public async setupTests(): Promise<void> {
    await execPromise(`rm -rf ${CLI_TEST_FOLDER}`);
    await execPromise(`rm -rf ${CLI_TEST_DUCKDB_FOLDER}`);

    // initially import 2 tables in source
    await execPromise(`${CLI_COMMAND} init ${CLI_TEST_DUCKDB_FOLDER_ARG}`);
    await execPromise(
      `${CLI_COMMAND} import-source test/data/AdBids.csv ${CLI_TEST_DUCKDB_FOLDER_ARG}`
    );
    await execPromise(
      `${CLI_COMMAND} import-source test/data/AdImpressions.tsv --name Impressions ${CLI_TEST_DUCKDB_FOLDER_ARG}`
    );
  }

  @FunctionalTestBase.Test()
  public async shouldLoadTablesFromDB() {
    await execPromise(
      `${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG} ` +
        `--db ${CLI_TEST_DUCKDB_FILE}`
    );
    let persistentState: PersistentTableState = JSON.parse(
      readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString()
    );
    let derivedState: DerivedTableState = JSON.parse(
      readFileSync(`${CLI_STATE_FOLDER}/derived_table_state.json`).toString()
    );
    console.log(persistentState.entities);
    expect(persistentState.entities[0].name).toBe("AdBids");
    this.assertColumns(derivedState.entities[0].profile, AdBidsColumnsTestData);
    expect(persistentState.entities[1].name).toBe("Impressions");
    this.assertColumns(
      derivedState.entities[1].profile,
      AdImpressionColumnsTestData
    );

    // drop a table and import another in source
    await execPromise(
      `${CLI_COMMAND} drop-source AdBids ${CLI_TEST_DUCKDB_FOLDER_ARG}`
    );
    await execPromise(
      `${CLI_COMMAND} import-source test/data/Users.csv ${CLI_TEST_DUCKDB_FOLDER_ARG}`
    );
    // trigger sync
    await execPromise(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);

    // verify tables are reflected in connected project
    persistentState = JSON.parse(
      readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString()
    );
    derivedState = JSON.parse(
      readFileSync(`${CLI_STATE_FOLDER}/derived_table_state.json`).toString()
    );
    expect(persistentState.entities[0].name).toBe("Impressions");
    this.assertColumns(
      derivedState.entities[0].profile,
      AdImpressionColumnsTestData
    );
    expect(persistentState.entities[1].name).toBe("Users");
    this.assertColumns(derivedState.entities[1].profile, UserColumnsTestData);

    // drop a table and import another in connected project
    await execPromise(
      `${CLI_COMMAND} drop-source Impressions ${CLI_TEST_FOLDER_ARG}`
    );
    await execPromise(
      `${CLI_COMMAND} import-source test/data/AdBids.csv ${CLI_TEST_FOLDER_ARG}`
    );
    // trigger sync
    await execPromise(`${CLI_COMMAND} init ${CLI_TEST_DUCKDB_FOLDER_ARG}`);

    // verify tables are reflected in source project
    // this happens without explicitly connecting during init
    persistentState = JSON.parse(
      readFileSync(
        `${CLI_STATE_DUCKDB_FOLDER}/persistent_table_state.json`
      ).toString()
    );
    expect(persistentState.entities[0].name).toBe("Users");
    expect(persistentState.entities[1].name).toBe("AdBids");
  }

  @FunctionalTestBase.Test()
  public async shouldUpdateProfilingData() {
    await execPromise(
      `${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG} ` +
        `--db ${CLI_TEST_DUCKDB_FILE}`
    );
    // update tables in a different function to auto close connection to db
    await execPromise(
      `ts-node-dev --project tsconfig.node.json -- ` +
        `test/utils/modify-db.ts ${CLI_TEST_DUCKDB_FILE}`
    );
    // trigger sync
    await execPromise(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);

    // verify tables are reflected in connected project
    const persistentState = JSON.parse(
      readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString()
    );
    const derivedState = JSON.parse(
      readFileSync(`${CLI_STATE_FOLDER}/derived_table_state.json`).toString()
    );
    console.log(persistentState.entities);
    expect(persistentState.entities[0].name).toBe("AdBids");
    this.assertColumns(derivedState.entities[0].profile, [
      ...AdBidsColumnsTestData.slice(0, 3),
      ...AdBidsColumnsTestData.slice(4),
    ]);
    expect(persistentState.entities[1].name).toBe("Impressions");
    this.assertColumns(derivedState.entities[1].profile, [
      ...AdImpressionColumnsTestData.slice(0, 2),
      {
        ...AdImpressionColumnsTestData[2],
        name: "r_country",
      },
      ...AdImpressionColumnsTestData.slice(3),
    ]);
    // temporary tables and views are not pulled
    expect(persistentState.entities.length).toBe(2);
  }

  @FunctionalTestBase.Test()
  public async shouldCopyDBToLocalDB() {
    await execPromise(
      `${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG} ` +
        `--db ${CLI_TEST_DUCKDB_FILE} --copy`
    );
    let persistentState: PersistentTableState = JSON.parse(
      readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString()
    );
    console.log(persistentState.entities);
    expect(persistentState.entities[0].name).toBe("AdBids");
    expect(persistentState.entities[1].name).toBe("Impressions");

    // drop a table and import another in source
    await execPromise(
      `${CLI_COMMAND} drop-source AdBids ${CLI_TEST_DUCKDB_FOLDER_ARG}`
    );
    await execPromise(
      `${CLI_COMMAND} import-source test/data/Users.csv ${CLI_TEST_DUCKDB_FOLDER_ARG}`
    );
    // trigger sync
    await execPromise(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);

    // verify tables are not updated in copied project
    persistentState = JSON.parse(
      readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString()
    );
    console.log(persistentState.entities);
    expect(persistentState.entities[0].name).toBe("AdBids");
    expect(persistentState.entities[1].name).toBe("Impressions");

    // drop a table and import another in copied project
    await execPromise(
      `${CLI_COMMAND} drop-source Impressions ${CLI_TEST_FOLDER_ARG}`
    );
    // Why does this statement hang!?
    // await execPromise(`${CLI_COMMAND} import-source test/data/AdBids.csv ${CLI_TEST_FOLDER_ARG}`);
    // trigger sync
    await execPromise(`${CLI_COMMAND} init ${CLI_TEST_DUCKDB_FOLDER_ARG}`);

    // verify tables are not touched in source project
    persistentState = JSON.parse(
      readFileSync(
        `${CLI_STATE_DUCKDB_FOLDER}/persistent_table_state.json`
      ).toString()
    );
    expect(persistentState.entities[0].name).toBe("Impressions");
    expect(persistentState.entities[1].name).toBe("Users");
  }
}
