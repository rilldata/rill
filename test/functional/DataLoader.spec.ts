import {TestBase} from "@adityahegde/typescript-test-utils";
import {FunctionalTestBase} from "./FunctionalTestBase";
import {
    ParquetFileTestData,
    ParquetFileTestDataProvider,
    TestDataColumns
} from "../data/DataLoader.data";
import {DATA_FOLDER} from "../data/generator/data-constants";
import {execSync} from "node:child_process";
import { assertColumns } from "./assertUtils";

@TestBase.Suite
export class DataLoaderSpec extends FunctionalTestBase {
    public parquetFileTestData(): ParquetFileTestDataProvider {
        return ParquetFileTestData;
    }

    @TestBase.Test("parquetFileTestData")
    public async shouldLoadParquetFile(parquetFile: string, cardinality: number, columns: TestDataColumns): Promise<void> {
        const actualFilePath = `${DATA_FOLDER}/${parquetFile}`;

        await this.clientDataModelerService.dispatch("addOrUpdateTable", [actualFilePath]);

        await this.waitForTables();

        const table = this.clientDataModelerStateService.getCurrentState().tables
            .find(tableFind => tableFind.path === actualFilePath);

        expect(table.path).toBe(actualFilePath);
        expect(table.cardinality).toBe(cardinality);

        assertColumns(table.profile, columns);
    }

    @TestBase.Test()
    public async shouldOnlyReloadNewFiles(): Promise<void> {
        await this.clientDataModelerService.dispatch("updateTablesFromSource", [DATA_FOLDER]);
        await this.waitForTables();

        const state = this.clientDataModelerStateService.getCurrentState();
        const adBidTable = state.tables.find(table => table.path.includes("AdBid"));
        const adImpressionTable = state.tables.find(table => table.path.includes("AdImpression"));

        execSync("touch data/AdBids.parquet");

        await this.clientDataModelerService.dispatch("updateTablesFromSource", [DATA_FOLDER]);
        await this.waitForTables();

        const newState = this.clientDataModelerStateService.getCurrentState();
        const newAdBidTable = newState.tables.find(table => table.path.includes("AdBid"));
        const newAdImpressionTable = newState.tables.find(table => table.path.includes("AdImpression"));

        expect(adBidTable.lastUpdated).toBeLessThan(newAdBidTable.lastUpdated);
        expect(adImpressionTable.lastUpdated).toBe(newAdImpressionTable.lastUpdated);
    }
}
