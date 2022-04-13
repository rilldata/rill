import type { DerivedTableState } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import type { PersistentTableState } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { exec } from "node:child_process";
import { readFileSync } from "fs";
import { AdBidsColumnsTestData, AdImpressionColumnsTestData } from "../data/DataLoader.data";
import {promisify} from "util"
import { FunctionalTestBase } from "./FunctionalTestBase";

const execPromise = promisify(exec);

const CLI_TEST_FOLDER = "temp/test-duckdb-import";
const CLI_STATE_FOLDER = `${CLI_TEST_FOLDER}/state`;
const DATA_MODELER_CLI = "npm run cli --";
const CLI_TEST_FOLDER_ARG = `--project ${CLI_TEST_FOLDER}`;

const CLI_TEST_DUCKDB_FOLDER = "temp/test-duckdb";
const CLI_TEST_DUCKDB_FILE = `${CLI_TEST_DUCKDB_FOLDER}/stage.db`;
const CLI_TEST_DUCKDB_FOLDER_ARG = `--project ${CLI_TEST_DUCKDB_FOLDER}`;

@FunctionalTestBase.Suite
export class DuckDbConnectionSpec extends FunctionalTestBase {
    @FunctionalTestBase.BeforeSuite()
    // override parent method to stop it from starting local server
    public async setup() {}

    @FunctionalTestBase.BeforeEachTest()
    public async setupTests(): Promise<void> {
        await execPromise(`rm -rf ${CLI_TEST_FOLDER}`);
        await execPromise(`rm -rf ${CLI_TEST_DUCKDB_FOLDER}`);
    }

    @FunctionalTestBase.Test()
    public async shouldLoadTablesFromDB() {
        await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_DUCKDB_FOLDER_ARG}`);
        await execPromise(`${DATA_MODELER_CLI} import-table data/AdBids.parquet ${CLI_TEST_DUCKDB_FOLDER_ARG}`);
        await execPromise(`${DATA_MODELER_CLI} import-table data/AdImpressions.parquet --name Impressions ${CLI_TEST_DUCKDB_FOLDER_ARG}`);

        await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_FOLDER_ARG} ` +
            `--duckdb ${CLI_TEST_DUCKDB_FILE}`);
        const persistentState: PersistentTableState =
            JSON.parse(readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString());
        const derivedState: DerivedTableState =
            JSON.parse(readFileSync(`${CLI_STATE_FOLDER}/derived_table_state.json`).toString());
        expect(persistentState.entities[0].name).toBe("AdBids");
        this.assertColumns(derivedState.entities[0].profile, AdBidsColumnsTestData);
        expect(persistentState.entities[1].name).toBe("Impressions");
        this.assertColumns(derivedState.entities[1].profile, AdImpressionColumnsTestData);
    }
}
