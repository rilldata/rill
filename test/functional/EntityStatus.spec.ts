import { FunctionalTestBase } from "./FunctionalTestBase";
import {
  EntityStatus,
  EntityType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { EntityStatusTracker } from "../utils/EntityStatusTracker";
import { asyncWait } from "$common/utils/waitUtils";
import { TwoSourceJoinQuery } from "../data/ModelQuery.data";
import { ApplicationStatus } from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";

@FunctionalTestBase.Suite
export class EntityStatusSpec extends FunctionalTestBase {
  private entityStatusTracker: EntityStatusTracker;

  @FunctionalTestBase.BeforeSuite()
  public async setupSuite() {
    this.entityStatusTracker = new EntityStatusTracker(
      this.serverDataModelerStateService,
      this.sandbox
    );
  }

  @FunctionalTestBase.BeforeEachTest()
  public async setupTests() {
    await this.clientDataModelerService.dispatch("clearAllSources", []);
    await this.clientDataModelerService.dispatch("clearAllModels", []);
    this.entityStatusTracker.init();
  }

  // @FunctionalTestBase.Test()
  // public async shouldHaveCorrectStatusWhileImportingSource() {
  //     this.entityStatusTracker.startTracker(EntityType.Source);
  //     await asyncWait(50);

  //     await this.clientDataModelerService.dispatch(
  //         "addOrUpdateSourceFromFile", ["test/data/AdBids.csv"]);
  //     await asyncWait(50);

  //     expect(this.entityStatusTracker.getStatusChangeOrder()).toEqual([
  //         EntityStatus.Importing,
  //         EntityStatus.Profiling,
  //         EntityStatus.Idle,
  //     ]);
  //     expect(this.entityStatusTracker.getApplicationStatusChangeOrder()).toEqual([
  //         ApplicationStatus.Idle,
  //         ApplicationStatus.Running,
  //         ApplicationStatus.Idle,
  //     ]);
  // }

  // @FunctionalTestBase.Test()
  // public async shouldHaveCorrectStatusWhileUpdatingModelQuery() {
  //     await this.clientDataModelerService.dispatch(
  //         "addOrUpdateSourceFromFile", ["test/data/AdBids.csv"]);
  //     await this.waitForSources();
  //     await this.clientDataModelerService.dispatch(
  //         "addModel", [{name: "query_0", query: ""}]);
  //     await asyncWait(50);

  //     const [model] = this.getModels("name", "query_0.sql");
  //     this.entityStatusTracker.startTracker(EntityType.Model);
  //     await asyncWait(50);

  //     await this.clientDataModelerService.dispatch(
  //         "updateModelQuery", [model.id, SingleSourceQuery]);
  //     await asyncWait(50);

  //     expect(this.entityStatusTracker.getStatusChangeOrder()).toEqual([
  //         EntityStatus.Idle,
  //         EntityStatus.Validating,
  //         EntityStatus.Profiling,
  //         EntityStatus.Idle,
  //     ]);
  //     expect(this.entityStatusTracker.getApplicationStatusChangeOrder()).toEqual([
  //         ApplicationStatus.Idle,
  //         ApplicationStatus.Running,
  //         ApplicationStatus.Idle,
  //     ]);
  // }

  @FunctionalTestBase.Test()
  public async shouldHaveCorrectStatusWhileExportingModel() {
    await this.clientDataModelerService.dispatch("addOrUpdateSourceFromFile", [
      "test/data/AdBids.csv",
    ]);
    await this.clientDataModelerService.dispatch("addOrUpdateSourceFromFile", [
      "test/data/AdImpressions.tsv",
    ]);
    await this.waitForSources();
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "query_0", query: TwoSourceJoinQuery },
    ]);
    await asyncWait(50);

    const [model] = this.getModels("name", "query_0.sql");
    this.entityStatusTracker.startTracker(EntityType.Model);
    await asyncWait(50);

    await this.clientDataModelerService.dispatch("exportToCsv", [
      model.id,
      "Joined.csv",
    ]);
    await asyncWait(50);

    expect(this.entityStatusTracker.getStatusChangeOrder()).toEqual([
      EntityStatus.Idle,
      EntityStatus.Exporting,
      EntityStatus.Idle,
    ]);
    expect(this.entityStatusTracker.getApplicationStatusChangeOrder()).toEqual([
      ApplicationStatus.Idle,
      ApplicationStatus.Running,
      ApplicationStatus.Idle,
    ]);
  }

  // @FunctionalTestBase.Test()
  // public async shouldOnlySwitchApplicationStatusOnce() {
  //     await this.clientDataModelerService.dispatch(
  //         "addModel", [{name: "query_0", query: ""}]);
  //     await this.clientDataModelerService.dispatch(
  //         "addModel", [{name: "query_1", query: ""}]);
  //     await asyncWait(50);

  //     const [model0] = this.getModels("name", "query_0.sql");
  //     const [model1] = this.getModels("name", "query_1.sql");

  //     this.entityStatusTracker.startTracker(EntityType.Source);
  //     await asyncWait(50);

  //     const promises = [];
  //     promises.push(this.clientDataModelerService.dispatch(
  //         "addOrUpdateSourceFromFile", ["test/data/AdBids.csv"]));
  //     promises.push(this.clientDataModelerService.dispatch(
  //         "addOrUpdateSourceFromFile", ["test/data/AdImpressions.tsv"]));
  //     await asyncWait(50);
  //     promises.push(this.clientDataModelerService.dispatch(
  //         "updateModelQuery", [model0.id, SingleSourceQuery]));
  //     promises.push(this.clientDataModelerService.dispatch(
  //         "updateModelQuery", [model1.id, TwoSourceJoinQuery]));
  //     await Promise.all(promises);
  //     await asyncWait(50);

  //     expect(this.entityStatusTracker.getApplicationStatusChangeOrder()).toEqual([
  //         ApplicationStatus.Idle,
  //         ApplicationStatus.Running,
  //         ApplicationStatus.Idle,
  //     ]);
  // }

  @FunctionalTestBase.AfterEachTest()
  public teardownTests() {
    this.entityStatusTracker.stopTracker();
  }
}
