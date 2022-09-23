import { ActionStatus } from "$web-local/common/data-modeler-service/response/ActionResponse";
import { ActionErrorType } from "$web-local/common/data-modeler-service/response/ActionResponseMessage";
import {
  EntityType,
  StateType,
} from "$web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { asyncWait } from "$web-local/common/utils/waitUtils";
import type { TestDataColumns } from "../data/DataLoader.data";
import {
  ModelQueryTestData,
  ModelQueryTestDataProvider,
  SingleTableQuery,
} from "../data/ModelQuery.data";
import { FunctionalTestBase } from "./FunctionalTestBase";

@FunctionalTestBase.Suite
export class ModelQuerySpec extends FunctionalTestBase {
  @FunctionalTestBase.BeforeSuite()
  public async setupTables(): Promise<void> {
    await this.loadTestTables();
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "model_0", query: "" },
    ]);
  }

  public modelQueryTestData(): ModelQueryTestDataProvider {
    return ModelQueryTestData;
  }

  @FunctionalTestBase.Test("modelQueryTestData")
  public async shouldUpdateModelQuery(
    query: string,
    columns: TestDataColumns
  ): Promise<void> {
    const [model] = this.getModels("tableName", "model_0");
    await this.clientDataModelerService.dispatch("updateModelQuery", [
      model.id,
      query,
    ]);
    await this.waitForModels();

    const [persistentModel, derivedModel] = this.getModels(
      "tableName",
      "model_0"
    );
    expect(derivedModel.error).toBeUndefined();
    expect(persistentModel.query).toBe(query);
    expect(derivedModel.cardinality).toBeGreaterThan(0);
    expect(derivedModel.preview.length).toBeGreaterThan(0);

    this.assertColumns(derivedModel.profile, columns);
  }

  @FunctionalTestBase.Test()
  public async shouldAddAndDeleteModel(): Promise<void> {
    const service = this.clientDataModelerStateService.getEntityStateService(
      EntityType.Model,
      StateType.Persistent
    );
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "newModel", query: SingleTableQuery },
    ]);

    await asyncWait(50);

    let newModel = service.getCurrentState().entities[1];
    expect(newModel.name).toBe("newModel.sql");

    const NEW_MODEL_UPDATE_NAME = "newModel_updated.sql";
    await this.clientDataModelerService.dispatch("updateModelName", [
      newModel.id,
      "newModel_updated",
    ]);
    await asyncWait(50);
    newModel = service.getCurrentState().entities[1];
    expect(newModel.name).toBe(NEW_MODEL_UPDATE_NAME);

    const OTHER_MODEL_NAME = "model_0.sql";
    await this.clientDataModelerService.dispatch("moveModelUp", [newModel.id]);
    await asyncWait(50);
    expect(service.getCurrentState().entities[0].name).toBe(
      NEW_MODEL_UPDATE_NAME
    );
    expect(service.getCurrentState().entities[1].name).toBe(OTHER_MODEL_NAME);

    await this.clientDataModelerService.dispatch("moveModelDown", [
      newModel.id,
    ]);
    await asyncWait(50);
    expect(service.getCurrentState().entities[0].name).toBe(OTHER_MODEL_NAME);
    expect(service.getCurrentState().entities[1].name).toBe(
      NEW_MODEL_UPDATE_NAME
    );

    await this.clientDataModelerService.dispatch("deleteModel", [newModel.id]);
    expect(service.getCurrentState().entities.length).toBe(1);
    expect(service.getCurrentState().entities[0].name).toBe(OTHER_MODEL_NAME);
  }

  @FunctionalTestBase.Test()
  public async shouldReturnModelQueryError(): Promise<void> {
    const INVALID_QUERY = "slect * from AdBids";

    const [model] = this.getModels("tableName", "model_0");
    // invalid query
    let response = await this.clientDataModelerService.dispatch(
      "updateModelQuery",
      [model.id, INVALID_QUERY]
    );
    await this.waitForModels();
    expect(response.status).toBe(ActionStatus.Failure);
    expect(response.messages[0].errorType).toBe(ActionErrorType.ModelQuery);
    let [, derivedModel] = this.getModels("tableName", "model_0");
    expect(derivedModel.error).toBe(response.messages[0].message);

    response = await this.clientDataModelerService.dispatch(
      "updateModelQuery",
      [model.id, INVALID_QUERY + " "]
    );
    await this.waitForModels();
    expect(response.status).toBe(ActionStatus.Failure);
    expect(response.messages[0].errorType).toBe(ActionErrorType.ModelQuery);
    [, derivedModel] = this.getModels("tableName", "model_0");
    expect(derivedModel.error).toBe(response.messages[0].message);

    // clearing query should clear the error
    response = await this.clientDataModelerService.dispatch(
      "updateModelQuery",
      [model.id, ""]
    );
    expect(response.status).toBe(ActionStatus.Success);
    expect(response.messages.length).toBe(0);
    [, derivedModel] = this.getModels("tableName", "model_0");
    expect(derivedModel.error).toBeUndefined();
  }

  @FunctionalTestBase.Test()
  public async shouldSwitchActiveEntityOnDelete() {
    for (let i = 1; i <= 5; i++) {
      await this.clientDataModelerService.dispatch("addModel", [
        { name: `model_${i}`, query: "" },
      ]);
      await asyncWait(50);
    }

    const ids = [];
    for (let i = 0; i < 5; i++) {
      const [model] = this.getModels("name", `model_${i}.sql`);
      ids.push(model.id);
    }

    await this.clientDataModelerService.dispatch("setActiveAsset", [
      EntityType.Model,
      ids[1],
    ]);
    await asyncWait(100);
    expect(this.getActiveEntity().id).toBe(ids[1]);

    await this.clientDataModelerService.dispatch("deleteModel", [ids[2]]);
    await asyncWait(100);
    expect(this.getActiveEntity().id).toBe(ids[1]);

    await this.clientDataModelerService.dispatch("deleteModel", [ids[1]]);
    await asyncWait(100);
    expect(this.getActiveEntity().id).toBe(ids[0]);

    await this.clientDataModelerService.dispatch("deleteModel", [ids[0]]);
    await asyncWait(100);
    expect(this.getActiveEntity().id).toBe(ids[3]);
  }

  @FunctionalTestBase.Test()
  public async shouldNotCreateOrRenameWithExistingTableName() {
    const resp = await this.clientDataModelerService.dispatch("addModel", [
      { name: "AdBids", query: "" },
    ]);
    await this.waitForModels();
    expect(resp.status).toBe(ActionStatus.Failure);
    expect(resp.messages[0].errorType).toBe(
      ActionErrorType.ExistingEntityError
    );

    await this.clientDataModelerService.dispatch("addModel", [
      { name: "modelRename", query: "" },
    ]);
    await this.waitForModels();

    const [model] = this.getModels("tableName", "modelRename");
    const renameResp = await this.clientDataModelerService.dispatch(
      "updateModelName",
      [model.id, "AdBids"]
    );
    expect(renameResp.status).toBe(ActionStatus.Failure);
    expect(renameResp.messages[0].errorType).toBe(
      ActionErrorType.ExistingEntityError
    );
  }
}
