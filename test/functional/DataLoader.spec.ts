import {TestBase} from "@adityahegde/typescript-test-utils";
import {FunctionalTestBase} from "./FunctionalTestBase";
import {
    ParquetFileTestData,
    ParquetFileTestDataProvider,
    TestDataColumns
} from "../data/DataLoader.data";

@TestBase.Suite
export class DataLoaderSpec extends FunctionalTestBase {
    public parquetFileTestData(): ParquetFileTestDataProvider {
        return ParquetFileTestData;
    }

    @TestBase.Test("parquetFileTestData")
    public async shouldLoadParquetFile(parquetFile: string, cardinality: number, columns: TestDataColumns): Promise<void> {
        await this.clientDataModelerService.dispatch("addOrUpdateDataset", [parquetFile]);

        await this.waitForDatasets();

        const dataset = this.clientDataModelerStateService.getCurrentState().sources
            .find(datasetFind => datasetFind.path === parquetFile);

        expect(dataset.path).toBe(parquetFile);
        expect(dataset.cardinality).toBe(cardinality);

        this.assertColumns(dataset.profile, columns);
    }
}
