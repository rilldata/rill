import {TestBase} from "@adityahegde/typescript-test-utils";
import {FunctionalTestBase} from "./FunctionalTestBase";
import {
    ParquetFileTestData, CSVFileTestData,
    FileImportTestDataProvider,
    TestDataColumns,
} from "../data/DataLoader.data";
import {DATA_FOLDER} from "../data/generator/data-constants";
import {execSync} from "node:child_process";
import { extractFileExtension, extractTableName } from "$lib/util/extract-table-name";

const AdBidsFile = "data/AdBids.parquet";
const AdImpressionsFile = "data/AdImpressions.parquet";

@TestBase.Suite
export class DataLoaderSpec extends FunctionalTestBase {
    public fileImportTestData(): FileImportTestDataProvider {
        return {
            subData: [ParquetFileTestData, CSVFileTestData],
        };
    }

    @TestBase.Test("fileImportTestData")
    public async shouldImportTableFromFile(inputFile: string, cardinality: number, columns: TestDataColumns): Promise<void> {
        const actualFilePath = `${DATA_FOLDER}/${inputFile}`;

        await this.clientDataModelerService.dispatch("addOrUpdateTableFromFile",
            [actualFilePath, `${extractTableName(inputFile)}_${extractFileExtension(inputFile)}`]);
        await this.waitForTables();

        const [table, derivedTable] = this.getTables("path", actualFilePath);

        expect(table.path).toBe(actualFilePath);
        expect(derivedTable.cardinality).toBe(cardinality);

        this.assertColumns(derivedTable.profile, columns);
    }

    @TestBase.Test()
    public async shouldOnlyReloadNewFiles(): Promise<void> {
        await this.clientDataModelerService.dispatch("addOrUpdateTableFromFile", [AdBidsFile]);
        await this.clientDataModelerService.dispatch("addOrUpdateTableFromFile", [AdImpressionsFile]);
        await this.waitForTables();

        const [adBidTable] = this.getTables("name", "AdBids");
        const [adImpressionTable] = this.getTables("name", "AdImpressions");

        execSync(`touch ${AdBidsFile}`);

        await this.clientDataModelerService.dispatch("addOrUpdateTableFromFile", [AdBidsFile]);
        await this.clientDataModelerService.dispatch("addOrUpdateTableFromFile", [AdImpressionsFile]);
        await this.waitForTables();

        const [newAdBidTable] = this.getTables("name", "AdBids");
        const [newAdImpressionTable] = this.getTables("name", "AdImpressions");

        expect(adBidTable.lastUpdated).toBeLessThan(newAdBidTable.lastUpdated);
        expect(adImpressionTable.lastUpdated).toBe(newAdImpressionTable.lastUpdated);
    }

    @TestBase.Test()
    public async shouldUseTableNameFromArgs(): Promise<void> {
        await this.clientDataModelerService.dispatch("addOrUpdateTableFromFile",
          [AdBidsFile, "AdBidsTable"]);
        await this.waitForTables();

        const [table] = this.getTables("name", "AdBidsTable");

        expect(table.path).toBe(AdBidsFile);
        expect(table.name).toBe("AdBidsTable");
    }

    @TestBase.Test()
    public async shouldNotLoadInvalidTable(): Promise<void> {
        await this.clientDataModelerService.dispatch("addOrUpdateTableFromFile",
          ["data/AdBids", "AdBidsTableInvalid"]);
        await this.waitForTables();

        const [table] = this.getTables("name", "AdBidsTableInvalid");

        expect(table).toBeUndefined();
    }
}
