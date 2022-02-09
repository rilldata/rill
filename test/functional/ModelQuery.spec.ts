import {FunctionalTestBase} from "./FunctionalTestBase";
import type {TestDataColumns} from "../data/DataLoader.data";
import {ModelQueryTestData, ModelQueryTestDataProvider, SingleTableQuery} from "../data/ModelQuery.data";
import {asyncWait} from "$common/utils/waitUtils";

@FunctionalTestBase.Suite
export class ModelQuerySpec extends FunctionalTestBase {
    @FunctionalTestBase.BeforeSuite()
    public async setupDataset(): Promise<void> {
        await this.loadTestTables();
    }

    public modelQueryTestData(): ModelQueryTestDataProvider {
        return ModelQueryTestData;
    }

    @FunctionalTestBase.Test("modelQueryTestData")
    public async shouldUpdateModelQuery(query: string, columns: TestDataColumns): Promise<void> {
        const modelId = this.clientDataModelerStateService.getCurrentState().queries[0].id;
        await this.clientDataModelerService.dispatch("updateModelQuery", [modelId, query]);
        await this.waitForModels();

        const model = this.clientDataModelerStateService.getCurrentState().queries[0];
        expect(model.error).toBeUndefined();
        expect(model.query).toBe(query);
        expect(model.cardinality).toBeGreaterThan(0);
        expect(model.preview.length).toBeGreaterThan(0);

        this.assertColumns(model.profile, columns);
    }

    @FunctionalTestBase.Test()
    public async shouldAddAndDeleteModel(): Promise<void> {
        await this.clientDataModelerService.dispatch("addModel",
            [{name: "newModel", query: SingleTableQuery}]);
        await asyncWait(50);

        let newModel = this.clientDataModelerStateService.getCurrentState().queries[1];
        expect(newModel.name).toBe("newModel.sql");

        const NEW_MODEL_UPDATE_NAME = "newModel_updated.sql";
        await this.clientDataModelerService.dispatch("updateModelName", [newModel.id, "newModel_updated"]);
        await asyncWait(50);
        newModel = this.clientDataModelerStateService.getCurrentState().queries[1];
        expect(newModel.name).toBe(NEW_MODEL_UPDATE_NAME);

        const OTHER_MODEL_NAME = "query_1.sql";
        await this.clientDataModelerService.dispatch("moveModelUp", [newModel.id]);
        await asyncWait(50);
        expect(this.clientDataModelerStateService.getCurrentState().queries[0].name).toBe(NEW_MODEL_UPDATE_NAME);
        expect(this.clientDataModelerStateService.getCurrentState().queries[1].name).toBe(OTHER_MODEL_NAME);

        await this.clientDataModelerService.dispatch("moveModelDown", [newModel.id]);
        await asyncWait(50);
        expect(this.clientDataModelerStateService.getCurrentState().queries[0].name).toBe(OTHER_MODEL_NAME);
        expect(this.clientDataModelerStateService.getCurrentState().queries[1].name).toBe(NEW_MODEL_UPDATE_NAME);

        await this.clientDataModelerService.dispatch("deleteModel", [newModel.id]);
        expect(this.clientDataModelerStateService.getCurrentState().queries.length).toBe(1);
        expect(this.clientDataModelerStateService.getCurrentState().queries[0].name).toBe(OTHER_MODEL_NAME);
    }
}
