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

const execPromise = promisify(exec);

const CLI_TEST_FOLDER = "temp/test-cli";
const CLI_STATE_FOLDER = `${CLI_TEST_FOLDER}/state`;
const DATA_MODELER_CLI = "./node_modules/.bin/ts-node-dev --project tsconfig.node.json -- src/cli/data-modeler-cli.ts";
const CLI_TEST_FOLDER_ARG = `--project ${CLI_TEST_FOLDER}`;

@FunctionalTestBase.Suite
export class CLISpec extends FunctionalTestBase {
    @FunctionalTestBase.BeforeSuite()
    // override parent method to stop it from starting local server
    public async setup() {}

    @FunctionalTestBase.BeforeEachTest()
    public async setupTests(): Promise<void> {
        await execPromise(`mkdir -p ${CLI_TEST_FOLDER}`);
        await execPromise(`rm -rf ${CLI_TEST_FOLDER}/*`);
    }

    // @FunctionalTestBase.Test()
    public async shouldInitProject(): Promise<void> {
        await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_FOLDER}`);
        expect(existsSync(`${CLI_TEST_FOLDER}/persistent_table_state.json`));
        expect(existsSync(`${CLI_TEST_FOLDER}/stage.db`));
    }

    // @FunctionalTestBase.Test()
    public async shouldAddTables(): Promise<void> {
        await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_FOLDER}`);
        await execPromise(`${DATA_MODELER_CLI} import-table data/AdBids.parquet ${CLI_TEST_FOLDER_ARG}`);
        await execPromise(`${DATA_MODELER_CLI} import-table data/AdImpressions.parquet --name Impressions ${CLI_TEST_FOLDER_ARG}`);

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
        await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_FOLDER}`);
        await execPromise(`${DATA_MODELER_CLI} import-table data/AdBids.parquet ${CLI_TEST_FOLDER_ARG}`);

        let persistentState: PersistentTableState =
            JSON.parse(readFileSync(`${CLI_STATE_FOLDER}/persistent_table_state.json`).toString());
        expect(persistentState.entities[0].name).toBe("AdBids");

        await execPromise(`${DATA_MODELER_CLI} drop-table AdBids ${CLI_TEST_FOLDER_ARG}`);

        persistentState = JSON.parse(readFileSync(
            `${CLI_STATE_FOLDER}/persistent_table_state.json`).toString());
        expect(persistentState.entities.length).toBe(0);
    }
}
