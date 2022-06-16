import { TestBase } from "@adityahegde/typescript-test-utils";
import { FunctionalTestBase } from "./FunctionalTestBase";
import {
  ParquetFileTestData,
  CSVFileTestData,
  FileImportTestDataProvider,
  TestDataColumns,
} from "../data/DataLoader.data";
import { DATA_FOLDER } from "../data/generator/data-constants";
import {
  extractFileExtension,
  extractSourceName,
} from "$lib/util/extract-source-name";
import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";
import { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";
import { SingleSourceQuery, TwoSourceJoinQuery } from "../data/ModelQuery.data";
import { asyncWait } from "$common/utils/waitUtils";

const UserFile = "test/data/Users.csv";

@TestBase.Suite
export class DataLoaderSpec extends FunctionalTestBase {
  @FunctionalTestBase.BeforeEachTest()
  public async setupTests() {
    await this.clientDataModelerService.dispatch("clearAllSources", []);
  }

  public fileImportTestData(): FileImportTestDataProvider {
    return {
      subData: [ParquetFileTestData, CSVFileTestData],
    };
  }

  // @TestBase.Test("fileImportTestData")
  public async shouldImportSourceFromFile(
    inputFile: string,
    cardinality: number,
    columns: TestDataColumns
  ): Promise<void> {
    const actualFilePath = `${DATA_FOLDER}/${inputFile}`;

    await this.clientDataModelerService.dispatch("addOrUpdateSourceFromFile", [
      actualFilePath,
      `${extractSourceName(inputFile)}_${extractFileExtension(inputFile)}`,
    ]);
    await this.waitForSources();
    await asyncWait(250);

    const [source, derivedSource] = this.getSources("path", actualFilePath);

    expect(source.path).toBe(actualFilePath);
    expect(derivedSource.cardinality).toBe(cardinality);

    this.assertColumns(derivedSource.profile, columns);
  }

  @TestBase.Test()
  public async shouldUseSourceNameFromArgs(): Promise<void> {
    await this.clientDataModelerService.dispatch("addOrUpdateSourceFromFile", [
      UserFile,
      "UsersSource",
    ]);
    await this.waitForSources();

    const [source] = this.getSources("name", "UsersSource");

    expect(source.path).toBe(UserFile);
    expect(source.name).toBe("UsersSource");
  }

  @TestBase.Test()
  public async shouldNotLoadInvalidSource(): Promise<void> {
    const response = await this.clientDataModelerService.dispatch(
      "addOrUpdateSourceFromFile",
      ["test/data/AdBids", "AdBidsSourceInvalid"]
    );
    await this.waitForSources();

    const [source] = this.getSources("name", "AdBidsSourceInvalid");

    expect(source).toBeUndefined();
    expect(response.status).toBe(ActionStatus.Failure);
    expect(response.messages[0].errorType).toBe(ActionErrorType.ImportSource);
  }

  @TestBase.Test()
  public async shouldDropSource(): Promise<void> {
    await Promise.all([
      this.clientDataModelerService.dispatch("addOrUpdateSourceFromFile", [
        "test/data/AdBids.csv",
      ]),
      this.clientDataModelerService.dispatch("addOrUpdateSourceFromFile", [
        "test/data/AdImpressions.csv",
      ]),
    ]);
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "query_0", query: SingleSourceQuery },
    ]);
    await this.waitForSources();
    await this.waitForModels();

    await this.clientDataModelerService.dispatch("dropSource", [
      "AdImpressions",
    ]);
    await asyncWait(100);

    const [source, derivedSource] = this.getSources(
      "sourceName",
      "AdImpressions"
    );
    expect(source).toBeUndefined();
    expect(derivedSource).toBeUndefined();

    const [model] = this.getModels("sourceName", "query_0");
    const response = await this.clientDataModelerService.dispatch(
      "updateModelQuery",
      [model.id, TwoSourceJoinQuery]
    );
    expect(response.status).toBe(ActionStatus.Failure);
  }

  @TestBase.Test()
  public async shouldNotImportWithExistingModelName() {
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "ExistingModel", query: "" },
    ]);
    await this.waitForModels();
    const resp = await this.clientDataModelerService.dispatch(
      "addOrUpdateSourceFromFile",
      ["test/data/AdBids.csv", "ExistingModel"]
    );
    await this.waitForSources();

    expect(resp.status).toBe(ActionStatus.Failure);
    expect(resp.messages[0].errorType).toBe(
      ActionErrorType.ExistingEntityError
    );
  }
}
