import { FunctionalTestBase } from "./FunctionalTestBase";
import { exec } from "node:child_process";
import {promisify} from "util"
import { existsSync, readFileSync } from "fs";
import type { DataModelerState } from "$lib/types";
import { AdBidsColumnsTestData, AdImpressionColumnsTestData } from "../data/DataLoader.data";

const execPromise = promisify(exec);

const CLI_TEST_FOLDER = "temp/test";
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

    @FunctionalTestBase.Test()
    public async shouldInitProject(): Promise<void> {
        await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_FOLDER}`);
        expect(existsSync(`${CLI_TEST_FOLDER}/saved-state.json`));
        expect(existsSync(`${CLI_TEST_FOLDER}/stage.db`));
    }

    @FunctionalTestBase.Test()
    public async shouldAddTables(): Promise<void> {
        await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_FOLDER}`);
        await execPromise(`${DATA_MODELER_CLI} import-table data/AdBids.parquet ${CLI_TEST_FOLDER_ARG}`);
        await execPromise(`${DATA_MODELER_CLI} import-table data/AdImpressions.parquet --name Impressions ${CLI_TEST_FOLDER_ARG}`);

        const state: DataModelerState = JSON.parse(readFileSync(`${CLI_TEST_FOLDER}/saved-state.json`).toString());
        expect(state.tables[0].name).toBe("AdBids");
        this.assertColumns(state.tables[0].profile, AdBidsColumnsTestData);
        expect(state.tables[1].name).toBe("Impressions");
        this.assertColumns(state.tables[1].profile, AdImpressionColumnsTestData);
    }
}
