import {DataModelerActions} from ".//DataModelerActions";
import type {DataModelerState, Dataset} from "$lib/types";
import {getNewDataset} from "$common/stateInstancesFactory";
import {ColumnarItemType} from "$common/data-modeler-state-service/ProfileColumnStateActions";
import {IDLE_STATUS, RUNNING_STATUS} from "$common/constants";
import {sanitizeTableName} from "$lib/util/sanitize-table-name";
import {getParquetFiles} from "$common/utils/getParquetFiles";
import {stat} from "fs/promises";

export class DatasetActions extends DataModelerActions {
    public async updateDatasetsFromSource(currentState: DataModelerState, sourcePath: string): Promise<void> {
        const files = await getParquetFiles(sourcePath);
        const filePaths = new Set(files);
        const newSources = currentState.sources.filter((dataset, index, self) => {
            if (!filePaths.has(dataset.path)) return false;
            return index === self.findIndex(indexCheckDataset => (indexCheckDataset.path === dataset.path));
        });
        if (currentState.sources.length !== newSources.length) {
            this.dataModelerStateService.dispatch("pruneAndDedupeDatasets", [files]);
        }

        await this.dataModelerService.dispatch("addOrUpdateAllDataset", [files]);
    }

    public async addOrUpdateAllDataset(currentState: DataModelerState, files: Array<string>): Promise<void> {
        const filePaths = new Set(files);
        await Promise.all(currentState.sources.map(async (dataset) => {
            const fileStats = await stat(dataset.path);
            if (fileStats.mtimeMs < dataset.lastUpdated) filePaths.delete(dataset.path);
            else filePaths.add(dataset.path);
        }));
        if (filePaths.size > 0) {
            await Promise.all([...filePaths].map(filePath =>
              this.dataModelerService.dispatch("addOrUpdateDataset", [filePath])));
        }
    }

    public async addOrUpdateDataset(currentState: DataModelerState, path: string): Promise<void> {
        const datasets = currentState.sources;
        const existingDataset = datasets.find(s => s.path === path);
        const dataset = {...(existingDataset || getNewDataset())};
        dataset.path = path;
        dataset.name = path.split("/").slice(-1)[0];
        dataset.tableName = sanitizeTableName(path);

        // get stats of the file and update only if it changed since we last saw it
        const fileStats = await stat(path);
        if (fileStats.mtimeMs < dataset.lastUpdated) return;
        dataset.lastUpdated = Date.now();

        this.dataModelerStateService.dispatch("addOrUpdateDatasetToState",
            [dataset, !existingDataset]);
        this.dataModelerStateService.dispatch("setDatasetStatus",
            [ColumnarItemType.Dataset, dataset.id, RUNNING_STATUS]);

        try {
            await this.collectDatasetInfo(dataset);

            await this.dataModelerService.dispatch("collectProfileColumns",
                [dataset.id, ColumnarItemType.Dataset]);
        } catch (err) {
            console.log(err);
        }

        this.dataModelerStateService.dispatch("setDatasetStatus",
            [ColumnarItemType.Dataset, dataset.id, IDLE_STATUS]);
    }

    // TODO: move this to something more meaningful
    public async setActiveAsset(currentState: DataModelerState, id: string, assetType: string): Promise<void> {
        this.dataModelerStateService.dispatch("setActiveAsset", [id, assetType]);
    }
    public async unsetActiveAsset(currentState: DataModelerState): Promise<void> {
        this.dataModelerStateService.dispatch("unsetActiveAsset", []);
    }

    private async collectDatasetInfo(dataset: Dataset) {
        await this.databaseService.dispatch("loadData", [dataset.path, dataset.tableName]);

        // create new dataset as one passed in args is readonly from the state.
        const newDataset: Dataset = {
            id: dataset.id,
            path: dataset.path,
            name: dataset.name,
            tableName: dataset.tableName,
            head: undefined,
        };

        await Promise.all([
            async () => {
                newDataset.profile = await this.databaseService.dispatch("getProfileColumns",
                    [dataset.tableName]);
                newDataset.profile = newDataset.profile
                    .filter(row => row.name !== "duckdb_schema" && row.name !== "schema" && row.name !== "root");
            },
            async () => newDataset.sizeInBytes =
                await this.databaseService.dispatch("getDestinationSize", [dataset.path]),
            async () => newDataset.cardinality =
                await this.databaseService.dispatch("getCardinalityOfTable", [dataset.tableName]),
            async () => newDataset.head =
                await this.databaseService.dispatch("getFirstNOfTable", [dataset.tableName]),
        ].map(asyncFunc => asyncFunc()));

        this.dataModelerStateService.dispatch("addOrUpdateDatasetToState",
            [newDataset, false]);
    }
}
