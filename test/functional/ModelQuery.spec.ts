import { FunctionalTestBase } from "./FunctionalTestBase";
import type { TestDataColumns } from "../data/DataLoader.data";
import { ModelQueryTestData, ModelQueryTestDataProvider, SingleTableQuery } from "../data/ModelQuery.data";
import { asyncWait } from "$common/utils/waitUtils";
import { EntityType, StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";
import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";

@FunctionalTestBase.Suite
export class ModelQuerySpec extends FunctionalTestBase {
    @FunctionalTestBase.BeforeSuite()
    public async setupTables(): Promise<void> {
        await this.loadTestTables();
        await this.clientDataModelerService.dispatch("addModel",
            [{name: "query_0", query: ""}]);
    }

    public modelQueryTestData(): ModelQueryTestDataProvider {
        return ModelQueryTestData;
    }

    @FunctionalTestBase.Test("modelQueryTestData")
    public async shouldUpdateModelQuery(query: string, columns: TestDataColumns): Promise<void> {
        const [model] = this.getModels("tableName", "query_0");
        await this.clientDataModelerService.dispatch("updateModelQuery",
            [model.id, query]);
        await this.waitForModels();

        const [persistentModel, derivedModel] = this.getModels("tableName", "query_0");
        expect(derivedModel.error).toBeUndefined();
        expect(persistentModel.query).toBe(query);
        expect(derivedModel.cardinality).toBeGreaterThan(0);
        expect(derivedModel.preview.length).toBeGreaterThan(0);

        this.assertColumns(derivedModel.profile, columns);
    }

    @FunctionalTestBase.Test()
    public async shouldAddAndDeleteModel(): Promise<void> {
        const service = this.clientDataModelerStateService
            .getEntityStateService(EntityType.Model, StateType.Persistent);
        await this.clientDataModelerService.dispatch("addModel",
            [{name: "newModel", query: SingleTableQuery}]);

        await asyncWait(50);

        let newModel = service.getCurrentState().entities[1];
        expect(newModel.name).toBe("newModel.sql");

        const NEW_MODEL_UPDATE_NAME = "newModel_updated.sql";
        await this.clientDataModelerService.dispatch("updateModelName", [newModel.id, "newModel_updated"]);
        await asyncWait(50);
        newModel = service.getCurrentState().entities[1];
        expect(newModel.name).toBe(NEW_MODEL_UPDATE_NAME);

        const OTHER_MODEL_NAME = "query_0.sql";
        await this.clientDataModelerService.dispatch("moveModelUp", [newModel.id]);
        await asyncWait(50);
        expect(service.getCurrentState().entities[0].name).toBe(NEW_MODEL_UPDATE_NAME);
        expect(service.getCurrentState().entities[1].name).toBe(OTHER_MODEL_NAME);

        await this.clientDataModelerService.dispatch("moveModelDown", [newModel.id]);
        await asyncWait(50);
        expect(service.getCurrentState().entities[0].name).toBe(OTHER_MODEL_NAME);
        expect(service.getCurrentState().entities[1].name).toBe(NEW_MODEL_UPDATE_NAME);

        await this.clientDataModelerService.dispatch("deleteModel", [newModel.id]);
        expect(service.getCurrentState().entities.length).toBe(1);
        expect(service.getCurrentState().entities[0].name).toBe(OTHER_MODEL_NAME);
    }

    @FunctionalTestBase.Test()
    public async shouldReturnModelQueryError(): Promise<void> {
        const INVALID_QUERY = "slect * from AdBids";

        const [model] = this.getModels("tableName", "query_0");
        // invalid query
        let response = await this.clientDataModelerService.dispatch("updateModelQuery",
            [model.id, INVALID_QUERY]);
        await this.waitForModels();
        expect(response.status).toBe(ActionStatus.Failure);
        expect(response.messages[0].errorType).toBe(ActionErrorType.ModelQuery);

        response = await this.clientDataModelerService.dispatch("updateModelQuery",
            [model.id, INVALID_QUERY + " "]);
        await this.waitForModels();
        expect(response.status).toBe(ActionStatus.Failure);
        expect(response.messages[0].errorType).toBe(ActionErrorType.ModelQuery);
    }
}
