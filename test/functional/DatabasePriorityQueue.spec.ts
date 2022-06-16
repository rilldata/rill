import { FunctionalTestBase } from "./FunctionalTestBase";
import { assert } from "sinon";
import {
  SingleSourceQuery,
  SingleSourceQueryColumnsTestData,
  TwoSourceJoinQuery,
  TwoSourceJoinQueryColumnsTestData,
} from "../data/ModelQuery.data";
import { asyncWait } from "$common/utils/waitUtils";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

@FunctionalTestBase.Suite
export class DatabasePriorityQueueSpec extends FunctionalTestBase {
  @FunctionalTestBase.BeforeEachTest()
  public async setupTests() {
    await this.clientDataModelerService.dispatch("clearAllSources", []);
    await this.clientDataModelerService.dispatch("clearAllModels", []);
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "query_0", query: "" },
    ]);
  }

  @FunctionalTestBase.Test()
  public async shouldDePrioritiseSourceProfiling() {
    const importPromise = this.clientDataModelerService.dispatch(
      "addOrUpdateSourceFromFile",
      ["test/data/AdBids.parquet"]
    );
    await asyncWait(1);

    const [model] = this.getModels("sourceName", "query_0");
    const modelQueryPromise = this.clientDataModelerService.dispatch(
      "updateModelQuery",
      [model.id, SingleSourceQuery]
    );

    await this.waitAndAssertPromiseOrder(modelQueryPromise, importPromise);
  }

  @FunctionalTestBase.Test()
  public async shouldStopOlderQueriesOfModel() {
    await this.clientDataModelerService.dispatch("addOrUpdateSourceFromFile", [
      "test/data/AdBids.parquet",
    ]);
    await this.clientDataModelerService.dispatch("addOrUpdateSourceFromFile", [
      "test/data/AdImpressions.parquet",
    ]);

    const [model] = this.getModels("sourceName", "query_0");
    const modelQueryOnePromise = this.clientDataModelerService.dispatch(
      "updateModelQuery",
      [model.id, TwoSourceJoinQuery]
    );
    await asyncWait(100);
    const modelQueryTwoPromise = this.clientDataModelerService.dispatch(
      "updateModelQuery",
      [model.id, SingleSourceQuery]
    );

    await this.waitAndAssertPromiseOrder(
      modelQueryOnePromise,
      modelQueryTwoPromise
    );
    const [, derivedModel] = this.getModels("sourceName", "query_0");
    this.assertColumns(derivedModel.profile, SingleSourceQueryColumnsTestData);
  }

  @FunctionalTestBase.Test()
  public async shouldDePrioritiseInactiveModel() {
    await this.clientDataModelerService.dispatch("addOrUpdateSourceFromFile", [
      "test/data/AdBids.parquet",
    ]);
    await this.clientDataModelerService.dispatch("addOrUpdateSourceFromFile", [
      "test/data/AdImpressions.parquet",
    ]);
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "query_1", query: "" },
    ]);

    const [model0] = this.getModels("sourceName", "query_1");
    const modelQueryOnePromise = this.clientDataModelerService.dispatch(
      "updateModelQuery",
      [model0.id, TwoSourceJoinQuery]
    );
    await this.clientDataModelerService.dispatch("setActiveAsset", [
      EntityType.Model,
      model0.id,
    ]);
    await asyncWait(50);
    const [model1] = this.getModels("sourceName", "query_0");
    const modelQueryTwoPromise = this.clientDataModelerService.dispatch(
      "updateModelQuery",
      [model1.id, SingleSourceQuery]
    );
    await asyncWait(50);
    await this.clientDataModelerService.dispatch("setActiveAsset", [
      EntityType.Model,
      model1.id,
    ]);

    await this.waitAndAssertPromiseOrder(
      modelQueryTwoPromise,
      modelQueryOnePromise
    );
  }

  @FunctionalTestBase.Test()
  public async shouldContinueModelProfileAfterAppendingSpaces() {
    await this.clientDataModelerService.dispatch("addOrUpdateSourceFromFile", [
      "test/data/AdImpressions.parquet",
    ]);

    const [model] = this.getModels("sourceName", "query_0");
    const modelQueryTwoPromise = this.clientDataModelerService.dispatch(
      "updateModelQuery",
      [model.id, TwoSourceJoinQuery]
    );
    await asyncWait(25);
    const modelQueryOnePromise = this.clientDataModelerService.dispatch(
      "updateModelQuery",
      [model.id, TwoSourceJoinQuery + "   \n"]
    );

    await this.waitAndAssertPromiseOrder(
      modelQueryOnePromise,
      modelQueryTwoPromise
    );
    const [, derivedModel] = this.getModels("sourceName", "query_0");
    this.assertColumns(derivedModel.profile, TwoSourceJoinQueryColumnsTestData);
  }

  private async waitAndAssertPromiseOrder(...promises: Array<Promise<any>>) {
    const spies = promises.map((promise) => {
      const spy = this.sandbox.spy();
      promise.then(spy);
      return spy;
    });

    await Promise.all(promises);
    assert.callOrder(...spies);
  }
}
