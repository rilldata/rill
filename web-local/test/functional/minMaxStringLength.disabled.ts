import type { DataProviderData } from "@adityahegde/typescript-test-utils";
import { expect } from "@jest/globals";
import { DATA_FOLDER } from "../data/generator/data-constants";
import { TwoTableJoinQuery } from "../data/ModelQuery.data";
import { FunctionalTestBase } from "./FunctionalTestBase";

@FunctionalTestBase.Suite
export class GetMinAndMaxStringLengthsTest extends FunctionalTestBase {
  @FunctionalTestBase.BeforeSuite()
  public async setupTables(): Promise<void> {
    await this.clientDataModelerService.dispatch("addOrUpdateTableFromFile", [
      `${DATA_FOLDER}/AdBids.csv`,
    ]);
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "model_0", query: "" },
    ]);
  }

  public minMaxStringLengthData(): DataProviderData<[string]> {
    return {
      subData: [
        {
          title: "happy path",
          args: [TwoTableJoinQuery],
        },
        {
          title: "double quote",
          args: ["select monthname(timestamp) from AdBids"],
        },
        {
          title: "single quote",
          args: ["select str_split('hello_world', '_')[1]"],
        },
        {
          title: "mixed quotes",
          args: ["select str_split(domain, '.')[1] from AdBids"],
        },
      ],
    };
  }

  @FunctionalTestBase.Test("minMaxStringLengthData")
  public async testSuccessfulQuery(query: string) {
    const [model] = this.getModels("tableName", "model_0");
    await this.clientDataModelerService.dispatch("updateModelQuery", [
      model.id,
      query,
    ]);
    await this.waitForModels();

    const [, derivedModel] = this.getModels("tableName", "model_0");
    for (const profile of derivedModel.profile) {
      expect(profile.largestStringLength).toBeGreaterThan(0);
    }
  }
}
