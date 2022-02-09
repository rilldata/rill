import {TestBase} from "@adityahegde/typescript-test-utils";
import {FunctionalTestBase} from "./FunctionalTestBase";
import {
    ParquetFileTestData,
    ParquetFileTestDataProvider,
    TestDataColumns
} from "../data/DataLoader.data";
import {DATA_FOLDER} from "../data/generator/data-constants";
import {execSync} from "node:child_process";

@TestBase.Suite
export class DataLoaderSpec extends FunctionalTestBase {
    public parquetFileTestData(): ParquetFileTestDataProvider {
        return ParquetFileTestData;
    }

    @TestBase.Test("parquetFileTestData")
    public async shouldLoadParquetFile(parquetFile: string, cardinality: number, columns: TestDataColumns): Promise<void> {
        const actualFilePath = `${DATA_FOLDER}/${parquetFile}`;

        await this.clientDataModelerService.dispatch("addOrUpdateDataset", [actualFilePath]);

        await this.waitForDatasets();

        const dataset = this.clientDataModelerStateService.getCurrentState().sources
            .find(datasetFind => datasetFind.path === actualFilePath);

        expect(dataset.path).toBe(actualFilePath);
        expect(dataset.cardinality).toBe(cardinality);

        this.assertColumns(dataset.profile, columns);
    }

    @TestBase.Test()
    public async shouldOnlyReloadNewFiles(): Promise<void> {
        await this.clientDataModelerService.dispatch("updateDatasetsFromSource", [DATA_FOLDER]);
        await this.waitForDatasets();

        const state = this.clientDataModelerStateService.getCurrentState();
        const adBidDataset = state.sources.find(dataset => dataset.path.includes("AdBid"));
        const adImpressionDataset = state.sources.find(dataset => dataset.path.includes("AdImpression"));

        execSync("touch data/AdBids.parquet");

        await this.clientDataModelerService.dispatch("updateDatasetsFromSource", [DATA_FOLDER]);
        await this.waitForDatasets();

        const newState = this.clientDataModelerStateService.getCurrentState();
        const newAdBidDataset = newState.sources.find(dataset => dataset.path.includes("AdBid"));
        const newAdImpressionDataset = newState.sources.find(dataset => dataset.path.includes("AdImpression"));

        expect(adBidDataset.lastUpdated).toBeLessThan(newAdBidDataset.lastUpdated);
        expect(adImpressionDataset.lastUpdated).toBe(newAdImpressionDataset.lastUpdated);
    }
}
