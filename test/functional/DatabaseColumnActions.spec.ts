import { FunctionalTestBase } from "./FunctionalTestBase";
import type { SinonSpy } from "sinon";
import { numericHistograms, dateHistograms } from "../data/HistogramSummary.data";

@FunctionalTestBase.Suite
export class DatabaseColumns extends FunctionalTestBase {
    private databaseDispatchSpy: SinonSpy;

    private async testHistogramSummary(input, output) {
        let [model] = this.getModels("tableName", "query_0");
        await this.clientDataModelerService.dispatch("updateModelQuery",
            [model.id, input]);
        await this.waitForModels();
        let [_, derivedModel] = this.getModels("tableName", "query_0");
        expect(derivedModel.profile[0].summary.histogram).toEqual(output);
    }

    @FunctionalTestBase.BeforeSuite()
    public async setupTables(): Promise<void> {
        await this.loadTestTables();
    }

    public async setup() {
        await super.setup();

        this.databaseDispatchSpy = this.sandbox.spy(
            this.serverDataModelerService.getDatabaseService(), "dispatch");
    }

    @FunctionalTestBase.BeforeEachTest()
    public async setupTests() {
        await this.clientDataModelerService.dispatch("clearAllModels", []);
        await this.clientDataModelerService.dispatch("addModel",
            [{name: "query_0", query: ""}]);
    }

    @FunctionalTestBase.Test()
    public async histogramsShouldComputeFromColumn() {
            for (const item of numericHistograms) {
                await this.testHistogramSummary(item.input, item.output);
            }
            for (const item of dateHistograms) {
                await this.testHistogramSummary(item.input, item.output);
            }
    }
}
