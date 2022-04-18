import { FunctionalTestBase } from "./FunctionalTestBase";
import { exec } from "node:child_process";
import {promisify} from "util"
import { existsSync, readFileSync } from "fs";
import { AdBidsColumnsTestData, AdImpressionColumnsTestData } from "../data/DataLoader.data";
import type {
    DerivedTableState
} from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import type {
    PersistentTableState
} from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import {CLI_COMMAND} from "../utils/getCliCommand";

const execPromise = promisify(exec);

const CLI_TEST_FOLDER = "temp/test-cli";
const CLI_STATE_FOLDER = `${CLI_TEST_FOLDER}/state`;
const CLI_TEST_FOLDER_ARG = `--project ${CLI_TEST_FOLDER}`;

@FunctionalTestBase.Suite
export class CLISpec extends FunctionalTestBase {
    @FunctionalTestBase.BeforeSuite()
    // override parent method to stop it from starting local server
    public async setup() {}

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
        await execPromise(`${CLI_COMMAND} import-table data/AdBids.parquet ${CLI_TEST_FOLDER_ARG}`);
        await execPromise(`${CLI_COMMAND} import-table data/AdImpressions.parquet --name Impressions ${CLI_TEST_FOLDER_ARG}`);

        const persistentState: PersistentTableState =
            JSON.parse(readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString());
        const derivedState: DerivedTableState =
            JSON.parse(readFileSync(`${CLI_STATE_FOLDER}/derived_table_state.json`).toString());
        expect(persistentState.entities[0].name).toBe("AdBids");
        this.assertColumns(derivedState.entities[0].profile, AdBidsColumnsTestData);
        expect(persistentState.entities[1].name).toBe("Impressions");
        this.assertColumns(derivedState.entities[1].profile, AdImpressionColumnsTestData);
    }

    @FunctionalTestBase.Test()
    public async shouldDropTable(): Promise<void> {
        await execPromise(`${CLI_COMMAND} init ${CLI_TEST_FOLDER_ARG}`);
        await execPromise(`${CLI_COMMAND} import-table data/AdBids.parquet ${CLI_TEST_FOLDER_ARG}`);

        let persistentState: PersistentTableState =
            JSON.parse(readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString());
        expect(persistentState.entities[0].name).toBe("AdBids");

        await execPromise(`${CLI_COMMAND} drop-table AdBids ${CLI_TEST_FOLDER_ARG}`);

        persistentState = JSON.parse(readFileSync(
            `${CLI_STATE_FOLDER}/persistent_table_state.json`).toString());
        expect(persistentState.entities.length).toBe(0);
    }
}
