import { expect } from "@jest/globals";
import { FunctionalTestBase } from "./FunctionalTestBase";
import { exec } from "node:child_process";
import { promisify } from "util";
import { existsSync, readFileSync } from "fs";
import {
  AdBidsColumnsTestData,
  AdImpressionColumnsTestData,
} from "../data/DataLoader.data";
import type { DerivedTableState } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import type { PersistentTableState } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { CLI_COMMAND } from "../utils/getCliCommand";

const execPromise = promisify(exec);
// uncomment this to better debug these tests
// const execVerbose = async (cmd) => {
//   const resp = await execPromise(cmd);
//   console.log(resp.stdout);
//   if (resp.stderr) console.log(resp.stderr);
// };
const execVerbose = execPromise;

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
    await execVerbose(`rm -rf ${CLI_TEST_FOLDER}`);
  }

  @FunctionalTestBase.Test()
  public async shouldInitProject(): Promise<void> {
    await execVerbose(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);
    expect(existsSync(`${CLI_TEST_FOLDER}/persistent_table_state.json`));
    expect(existsSync(`${CLI_TEST_FOLDER}/stage.db`));
  }

  @FunctionalTestBase.Test()
  public async shouldAddTables(): Promise<void> {
    await execVerbose(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);
    await execVerbose(
      `${CLI_COMMAND} import-source test/data/AdBids.csv ${CLI_TEST_FOLDER_ARG}`
    );
    await execVerbose(
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

  // @FunctionalTestBase.Test()
  // public async shouldErrorIfSourceFileIsMalformed(): Promise<void> {
  //   await execPromise(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);
  //   await execPromise(
  //     `${CLI_COMMAND} import-source test/data/AdBids.parquet ${CLI_TEST_FOLDER_ARG}`
  //   );
  //   // import the broken dataset.
  //   await execVerbose(
  //     `${CLI_COMMAND} import-source test/data/BrokenCSV.csv ${CLI_TEST_FOLDER_ARG}`
  //   );

  //   let persistentState: PersistentTableState = JSON.parse(
  //     readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString()
  //   );

  //   // BrokenCSV should not be present in the state.
  //   const brokenCSVState = persistentState.entities.find(
  //     (entity) => entity.tableName === "BrokenCSV"
  //   );
  //   expect(brokenCSVState).toBeUndefined();

  //   // let's get the state for AdBids before we attempt to import a broken dataset into it.
  //   const adBids = persistentState.entities.find(
  //     (entity) => entity.tableName === "AdBids"
  //   );
  //   // let's try to replace AdBids
  //   await execVerbose(
  //     `${CLI_COMMAND} import-source test/data/BrokenCSV.csv --name AdBids --force ${CLI_TEST_FOLDER_ARG}`
  //   );
  //   // check to see if the sources are the same.
  //   persistentState = JSON.parse(
  //     readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString()
  //   );

  //   const newAdBids = persistentState.entities.find(
  //     (entity) => entity.tableName === "AdBids"
  //   );
  //   const oldStateObject = { ...adBids };

  //   // check the newAdBids
  //   expect(newAdBids).toEqual(expect.objectContaining(oldStateObject));
  // }

  // @FunctionalTestBase.Test()
  // public async shouldDropTable(): Promise<void> {
  //   await execVerbose(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);
  //   await execVerbose(
  //     `${CLI_COMMAND} import-source test/data/AdBids.parquet ${CLI_TEST_FOLDER_ARG}`
  //   );

  //   let persistentState: PersistentTableState = JSON.parse(
  //     readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString()
  //   );
  //   expect(persistentState.entities[0].name).toBe("AdBids");

  //   await execVerbose(
  //     `${CLI_COMMAND} drop-source AdBids ${CLI_TEST_FOLDER_ARG}`
  //   );

  //   persistentState = JSON.parse(
  //     readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString()
  //   );
  //   expect(persistentState.entities.length).toBe(0);
  // }

  // @FunctionalTestBase.Test()
  // public async shouldInitDbAtTheBeginning(): Promise<void> {
  //   await execVerbose(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);
  //   await execVerbose(
  //     `${CLI_COMMAND} import-source test/data/AdBids.csv ${CLI_TEST_FOLDER_ARG}`
  //   );

  //   const cp = exec(`${CLI_COMMAND} start ${CLI_TEST_FOLDER_ARG}`);
  //   writeFileSync(
  //     `${CLI_TEST_FOLDER}/models/TestBroken.sql`,
  //     "SELECT * FROM AdBids"
  //   );
  //   await asyncWaitUntil(() => isPortOpen(8080));
  //   treeKill(cp.pid);

  //   const persistentModelState: PersistentModelState = JSON.parse(
  //     readFileSync(`${CLI_STATE_FOLDER}/persistent_model_state.json`).toString()
  //   );
  //   const model = persistentModelState.entities.find(
  //     (m) => m.tableName === "TestBroken"
  //   );
  //   const derivedModelState: DerivedModelState = JSON.parse(
  //     readFileSync(`${CLI_STATE_FOLDER}/derived_model_state.json`).toString()
  //   );
  //   const derivedModel = derivedModelState.entities.find(
  //     (d) => d.id === model.id
  //   );
  //   expect(derivedModel.error).toBeUndefined();
  // }
}
