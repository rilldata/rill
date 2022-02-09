import {DataModelerActions} from ".//DataModelerActions";
import type {DataModelerState, Dataset} from "$lib/types";
import {newSource} from "$common/dataFactory";
import {ColumnarItemType} from "$common/data-modeler-state-service/ProfileColumnStateActions";
import {IDLE_STATUS, RUNNING_STATUS} from "$common/constants";
import {sanitizeTableName} from "$lib/util/sanitize-table-name";
import {getParquetFiles} from "$common/utils/getParquetFiles";
import {stat} from "fs/promises";

export class DatasetActions extends DataModelerActions {
    public async clearDatasets(currentState: DataModelerState): Promise<void> {
        // TODO
    }

    public async updateDatasetsFromSource(currentState: DataModelerState, sourcePath: string): Promise<void> {
        const files = await getParquetFiles(sourcePath);
        this.dataModelerStateService.dispatch("pruneAndDedupeDatasets", [files]);
        await Promise.all(files.map(file => this.dataModelerService.dispatch("addOrUpdateDataset", [file])));
    }

    public async addOrUpdateDataset(currentState: DataModelerState, path: string): Promise<void> {
        const datasets = currentState.sources;
        const existingDataset = datasets.find(s => s.path === path);
        const dataset = {...(existingDataset || newSource())};
        dataset.path = path;
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
