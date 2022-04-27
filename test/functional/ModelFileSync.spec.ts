import {FunctionalTestBase} from "./FunctionalTestBase";
import {RootConfig} from "$common/config/RootConfig";
import {DatabaseConfig} from "$common/config/DatabaseConfig";
import {StateConfig} from "$common/config/StateConfig";
import {
    SingleTableQuery,
    SingleTableQueryColumnsTestData,
    TwoTableJoinQuery,
    TwoTableJoinQueryColumnsTestData
} from "../data/ModelQuery.data";
import {existsSync, readFileSync, writeFileSync} from "fs";
import {asyncWait} from "$common/utils/waitUtils";
import {expect} from "@playwright/test";
import {execSync} from "node:child_process";

const SYNC_TEST_FOLDER = "temp/model-sync-test";
const MODEL_FOLDER = `${SYNC_TEST_FOLDER}/models`;

@FunctionalTestBase.Suite
export class ModelFileSyncSpec extends FunctionalTestBase {
    @FunctionalTestBase.BeforeSuite()
    public async setupTables(): Promise<void> {
        await this.loadTestTables();
    }

    public async setup() {
        execSync(`rm -rf ${SYNC_TEST_FOLDER}`);
        await super.setup(new RootConfig({
            database: new DatabaseConfig({ databaseName: ":memory:" }),
            state: new StateConfig({ autoSync: true, syncInterval: 50 }),
            projectFolder: SYNC_TEST_FOLDER, profileWithUpdate: false,
        }));
    }

    @FunctionalTestBase.Test()
    public async shouldUpdateModelFileAndViceVersa() {
        const QUERY_FILE = `${MODEL_FOLDER}/query_0.sql`;

        expect(existsSync(QUERY_FILE)).toBe(false);

        await this.clientDataModelerService.dispatch("addModel",
            [{name: "query_0", query: ""}]);
        await asyncWait(100);
        expect(existsSync(QUERY_FILE)).toBe(true);

        const [model, ] = this.getModels("tableName", "query_0");
        await this.clientDataModelerService.dispatch("updateModelQuery",
            [model.id, SingleTableQuery]);
        await asyncWait(100);

        // updating query from client should update the file
        expect(readFileSync(QUERY_FILE).toString()).toBe(SingleTableQuery);
        const [, persistentModel] = this.getModels("tableName", "query_0");
        this.assertColumns(persistentModel.profile, SingleTableQueryColumnsTestData);

        // updating query from file should update profiling data
        writeFileSync(QUERY_FILE, TwoTableJoinQuery);
        await asyncWait(100);
        const [, newPersistentModel] = this.getModels("tableName", "query_0");
        this.assertColumns(newPersistentModel.profile, TwoTableJoinQueryColumnsTestData);
    }

    @FunctionalTestBase.Test()
    public async shouldRecreateModelFileOnDelete() {
        const QUERY_FILE = `${MODEL_FOLDER}/query_1.sql`;

        await this.clientDataModelerService.dispatch("addModel",
            [{name: "query_1", query: SingleTableQuery}]);
        await asyncWait(100);
        expect(existsSync(QUERY_FILE)).toBe(true);

        // file is recreated if deleted.
        execSync(`rm ${QUERY_FILE}`);
        await asyncWait(100);
        const [model, ] = this.getModels("tableName", "query_1");
        expect(model.query).toBe(SingleTableQuery);
        expect(existsSync(QUERY_FILE)).toBe(true);
    }
}
