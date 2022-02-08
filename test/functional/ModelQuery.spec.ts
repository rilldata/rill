import {FunctionalTestBase} from "./FunctionalTestBase";
import {ParquetFileTestData, TestDataColumns} from "../data/DataLoader.data";
import {QueryInfoTestData, QueryInfoTestDataProvider} from "../data/ModelQuery.data";

@FunctionalTestBase.Suite
export class ModelQuerySpec extends FunctionalTestBase {
    @FunctionalTestBase.BeforeSuite()
    public async setupDataset(): Promise<void> {
        await Promise.all(ParquetFileTestData.subData.map(async (parquetFileData) => {
            await this.clientDataModelerActionAPI.dispatch("addOrUpdateDataset", [parquetFileData.title]);
        }));
        await this.waitForDatasets();
    }

    public queryInfoTestData(): QueryInfoTestDataProvider {
        return QueryInfoTestData;
    }

    @FunctionalTestBase.Test("queryInfoTestData")
    public async shouldUpdateQueryInfo(query: string, columns: TestDataColumns): Promise<void> {
        const modelId = this.clientDataModelerStateManager.getCurrentState().queries[0].id;
        await this.clientDataModelerActionAPI.dispatch("updateQueryInformation", [modelId, query]);
        await this.waitForModels();

        const model = this.clientDataModelerStateManager.getCurrentState().queries[0];
        expect(model.error).toBeUndefined();
        expect(model.query).toBe(query);
        expect(model.cardinality).toBeGreaterThan(0);
        expect(model.preview.length).toBeGreaterThan(0);

        this.assertColumns(model.profile, columns);
    }
}
