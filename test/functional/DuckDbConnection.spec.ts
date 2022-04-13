import type { DerivedTableState } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import type { PersistentTableState } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { exec } from "node:child_process";
import { readFileSync } from "fs";
import { AdBidsColumnsTestData, AdImpressionColumnsTestData, UserColumnsTestData } from "../data/DataLoader.data";
import {promisify} from "util";
import { FunctionalTestBase } from "./FunctionalTestBase";
import {getCliCommand} from "../utils/getCliCommand";

const execPromise = promisify(exec);

const CLI_TEST_FOLDER = "temp/test-duckdb-import";
const CLI_STATE_FOLDER = `${CLI_TEST_FOLDER}/state`;
const DATA_MODELER_CLI = getCliCommand();
const CLI_TEST_FOLDER_ARG = `--project ${CLI_TEST_FOLDER}`;

const CLI_TEST_DUCKDB_FOLDER = "temp/test-duckdb";
const CLI_STATE_DUCKDB_FOLDER = `${CLI_TEST_DUCKDB_FOLDER}/state`;
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

        // initially import 2 tables in source
        await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_DUCKDB_FOLDER_ARG}`);
        await execPromise(`${DATA_MODELER_CLI} import-table data/AdBids.parquet ${CLI_TEST_DUCKDB_FOLDER_ARG}`);
        await execPromise(`${DATA_MODELER_CLI} import-table data/AdImpressions.parquet --name Impressions ${CLI_TEST_DUCKDB_FOLDER_ARG}`);
    }

    @FunctionalTestBase.Test()
    public async shouldLoadTablesFromDB() {
        await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_FOLDER_ARG} ` +
            `--db ${CLI_TEST_DUCKDB_FILE}`);
        let persistentState: PersistentTableState = JSON.parse(readFileSync(
            `${CLI_STATE_FOLDER}/persistent_table_state.json`).toString());
        let derivedState: DerivedTableState = JSON.parse(readFileSync(
            `${CLI_STATE_FOLDER}/derived_table_state.json`).toString());
        expect(persistentState.entities[0].name).toBe("AdBids");
        this.assertColumns(derivedState.entities[0].profile, AdBidsColumnsTestData);
        expect(persistentState.entities[1].name).toBe("Impressions");
        this.assertColumns(derivedState.entities[1].profile, AdImpressionColumnsTestData);

        // drop a table and import another in source
        await execPromise(`${DATA_MODELER_CLI} drop-table AdBids ${CLI_TEST_DUCKDB_FOLDER_ARG}`);
        await execPromise(`${DATA_MODELER_CLI} import-table data/Users.csv ${CLI_TEST_DUCKDB_FOLDER_ARG}`);
        // trigger sync
        await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_FOLDER_ARG}`);

        // verify tables are reflected in connected project
        persistentState = JSON.parse(readFileSync(
            `${CLI_STATE_FOLDER}/persistent_table_state.json`).toString());
        derivedState = JSON.parse(readFileSync(
            `${CLI_STATE_FOLDER}/derived_table_state.json`).toString());
        expect(persistentState.entities[0].name).toBe("Impressions");
        this.assertColumns(derivedState.entities[0].profile, AdImpressionColumnsTestData);
        expect(persistentState.entities[1].name).toBe("Users");
        this.assertColumns(derivedState.entities[1].profile, UserColumnsTestData);

        // drop a table and import another in connected project
        await execPromise(`${DATA_MODELER_CLI} drop-table Impressions ${CLI_TEST_FOLDER_ARG}`);
        await execPromise(`${DATA_MODELER_CLI} import-table data/AdBids.csv ${CLI_TEST_FOLDER_ARG}`);
        // trigger sync
        await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_DUCKDB_FOLDER_ARG}`);

        // verify tables are reflected in source project
        // this happens without explicitly connecting during init
        persistentState = JSON.parse(readFileSync(
            `${CLI_STATE_DUCKDB_FOLDER}/persistent_table_state.json`).toString());
        expect(persistentState.entities[0].name).toBe("Users");
        expect(persistentState.entities[1].name).toBe("AdBids");
    }

    @FunctionalTestBase.Test()
    public async shouldCopyDBToLocalDB() {
        await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_FOLDER_ARG} ` +
            `--db ${CLI_TEST_DUCKDB_FILE} --copy`);
        let persistentState: PersistentTableState = JSON.parse(readFileSync(
            `${CLI_STATE_FOLDER}/persistent_table_state.json`).toString());
        expect(persistentState.entities[0].name).toBe("AdBids");
        expect(persistentState.entities[1].name).toBe("Impressions");

        // drop a table and import another in source
        await execPromise(`${DATA_MODELER_CLI} drop-table AdBids ${CLI_TEST_DUCKDB_FOLDER_ARG}`);
        await execPromise(`${DATA_MODELER_CLI} import-table data/Users.csv ${CLI_TEST_DUCKDB_FOLDER_ARG}`);
        // trigger sync
        await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_FOLDER_ARG}`);

        // verify tables are not updated in copied project
        persistentState = JSON.parse(readFileSync(
            `${CLI_STATE_FOLDER}/persistent_table_state.json`).toString());
        expect(persistentState.entities[0].name).toBe("AdBids");
        expect(persistentState.entities[1].name).toBe("Impressions");

        // drop a table and import another in copied project
        await execPromise(`${DATA_MODELER_CLI} drop-table Impressions ${CLI_TEST_FOLDER_ARG}`);
        // Why does this statement hang!?
        // await execPromise(`${DATA_MODELER_CLI} import-table data/AdBids.csv ${CLI_TEST_FOLDER_ARG}`);
        // trigger sync
        await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_DUCKDB_FOLDER_ARG}`);

        // verify tables are not touched in source project
        persistentState = JSON.parse(readFileSync(
            `${CLI_STATE_DUCKDB_FOLDER}/persistent_table_state.json`).toString());
        expect(persistentState.entities[0].name).toBe("Impressions");
        expect(persistentState.entities[1].name).toBe("Users");
    }
}
