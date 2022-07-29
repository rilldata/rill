import type { DerivedTableState } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import type { PersistentTableState } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { exec } from "node:child_process";
import { readFileSync } from "fs";
import {
  AdBidsColumnsTestData,
  AdImpressionColumnsTestData,
  TestDataColumns,
  UserColumnsTestData,
} from "../data/DataLoader.data";
import { promisify } from "util";
import { FunctionalTestBase } from "./FunctionalTestBase";
import { CLI_COMMAND } from "../utils/getCliCommand";

const execPromise = promisify(exec);
// uncomment this to better debug these tests
// const execVerbose = async (cmd) => {
//   const resp = await execPromise(cmd);
//   console.log(resp.stdout);
//   if (resp.stderr) console.log(resp.stderr);
// };
const execVerbose = execPromise;

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
    await execVerbose(`rm -rf ${CLI_TEST_FOLDER}`);
    await execVerbose(`rm -rf ${CLI_TEST_DUCKDB_FOLDER}`);

    // initially import 2 tables in source
    await execVerbose(`${CLI_COMMAND} init ${CLI_TEST_DUCKDB_FOLDER_ARG}`);
    await execVerbose(
      `${CLI_COMMAND} import-source test/data/AdBids.csv ${CLI_TEST_DUCKDB_FOLDER_ARG}`
    );
    await execVerbose(
      `${CLI_COMMAND} import-source test/data/AdImpressions.tsv --name Impressions ${CLI_TEST_DUCKDB_FOLDER_ARG}`
    );
  }

  // @FunctionalTestBase.Test()
  public async shouldLoadTablesFromDB() {
    await execVerbose(
      `${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG} ` +
        `--db ${CLI_TEST_DUCKDB_FILE}`
    );
    this.assertTables(
      CLI_STATE_FOLDER,
      ["AdBids", "Impressions"],
      [AdBidsColumnsTestData, AdImpressionColumnsTestData]
    );

    // drop a table and import another in source
    await execVerbose(
      `${CLI_COMMAND} drop-source AdBids ${CLI_TEST_DUCKDB_FOLDER_ARG}`
    );
    await execVerbose(
      `${CLI_COMMAND} import-source test/data/Users.csv ${CLI_TEST_DUCKDB_FOLDER_ARG}`
    );
    // trigger sync
    await execVerbose(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);

    // verify tables are reflected in connected project
    this.assertTables(
      CLI_STATE_FOLDER,
      ["Impressions", "Users"],
      [AdImpressionColumnsTestData, UserColumnsTestData]
    );

    // drop a table and import another in connected project
    await execVerbose(
      `${CLI_COMMAND} drop-source Impressions ${CLI_TEST_FOLDER_ARG}`
    );
    await execVerbose(
      `${CLI_COMMAND} import-source test/data/AdBids.csv ${CLI_TEST_FOLDER_ARG}`
    );
    // trigger sync
    await execVerbose(`${CLI_COMMAND} init ${CLI_TEST_DUCKDB_FOLDER_ARG}`);

    // verify tables are reflected in source project
    // this happens without explicitly connecting during init
    this.assertTables(CLI_STATE_DUCKDB_FOLDER, ["AdBids", "Users"]);
  }

  @FunctionalTestBase.Test()
  public async shouldUpdateProfilingData() {
    await execVerbose(
      `${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG} ` +
        `--db ${CLI_TEST_DUCKDB_FILE}`
    );
    // update tables in a different function to auto close connection to database
    await execVerbose(
      `ts-node-dev --project tsconfig.node.json -- ` +
        `test/utils/modify-db.ts ${CLI_TEST_DUCKDB_FILE}`
    );
    // trigger sync
    await execVerbose(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);

    // verify tables are reflected in connected project
    this.assertTables(
      CLI_STATE_FOLDER,
      ["AdBids", "Impressions"],
      [
        [
          ...AdBidsColumnsTestData.slice(0, 3),
          ...AdBidsColumnsTestData.slice(4),
        ],
        [
          ...AdImpressionColumnsTestData.slice(0, 2),
          {
            ...AdImpressionColumnsTestData[2],
            name: "r_country",
          },
          ...AdImpressionColumnsTestData.slice(3),
        ],
      ]
    );
  }

  // @FunctionalTestBase.Test()
  public async shouldCopyDBToLocalDB() {
    await execVerbose(
      `${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG} ` +
        `--db ${CLI_TEST_DUCKDB_FILE} --copy`
    );
    this.assertTables(CLI_STATE_FOLDER, ["AdBids", "Impressions"]);

    // drop a table and import another in source
    await execVerbose(
      `${CLI_COMMAND} drop-source AdBids ${CLI_TEST_DUCKDB_FOLDER_ARG}`
    );
    await execVerbose(
      `${CLI_COMMAND} import-source test/data/Users.csv ${CLI_TEST_DUCKDB_FOLDER_ARG}`
    );
    // trigger sync
    await execVerbose(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);

    // verify tables are not updated in copied project
    this.assertTables(CLI_STATE_FOLDER, ["AdBids", "Impressions"]);

    // drop a table and import another in copied project
    await execVerbose(
      `${CLI_COMMAND} drop-source Impressions ${CLI_TEST_FOLDER_ARG}`
    );
    // Why does this statement hang!?
    // await execVerbose(`${CLI_COMMAND} import-source test/data/AdBids.csv ${CLI_TEST_FOLDER_ARG}`);
    // trigger sync
    await execVerbose(`${CLI_COMMAND} init ${CLI_TEST_DUCKDB_FOLDER_ARG}`);

    // verify tables are not touched in source project
    this.assertTables(CLI_STATE_DUCKDB_FOLDER, ["Impressions", "Users"]);
  }

  private assertTables(
    stateFolder: string,
    tableNames: Array<string>,
    tableColumns?: Array<TestDataColumns>
  ) {
    const persistentState: PersistentTableState = JSON.parse(
      readFileSync(`${stateFolder}/persistent_table_state.json`).toString()
    );
    const derivedState: DerivedTableState = tableColumns
      ? JSON.parse(
          readFileSync(`${stateFolder}/derived_table_state.json`).toString()
        )
      : {};

    expect(persistentState.entities.length).toBe(tableNames.length);
    tableNames.forEach((tableName, index) => {
      const persistentTable = persistentState.entities.find(
        (table) => table.tableName === tableName
      );
      expect(persistentTable).not.toBeUndefined();
      if (tableColumns) {
        const derivedTable = derivedState.entities.find(
          (table) => table.id === persistentTable.id
        );
        this.assertColumns(derivedTable.profile, tableColumns[index]);
      }
    });
  }
}
