import {DataModelerActions} from "../data-modeler-actions/DataModelerActions";
import type {DataModelerState, Dataset} from "$lib/types";
import {newSource} from "$common/data-factory";
import {ColumnarItemType} from "$common/state-actions/ProfileColumnStateActions";
import {IDLE_STATUS, RUNNING_STATUS} from "$common/constants";
import {sanitizeTableName} from "$lib/util/sanitize-table-name";

export class DatasetActions extends DataModelerActions {
    public async addOrUpdateDataset(currentState: DataModelerState, path: string): Promise<void> {
        const datasets = currentState.sources;
        const existingDataset = datasets.find(s => s.path === path);
        const dataset = {...(existingDataset || newSource())};
        dataset.path = path;
        dataset.tableName = sanitizeTableName(path);

        this.dataModelerStateManager.dispatch("addOrUpdateDatasetToState",
            [dataset, !existingDataset]);
        this.dataModelerStateManager.dispatch("setDatasetStatus",
            [ColumnarItemType.Dataset, dataset.id, RUNNING_STATUS]);

        try {
            await this.collectDatasetInfo(dataset);

            await this.dataModelerActionAPI.dispatch("collectProfileColumns",
                [dataset.id, ColumnarItemType.Dataset]);
        } catch (err) {
            console.log(err);
        }

        this.dataModelerStateManager.dispatch("setDatasetStatus",
            [ColumnarItemType.Dataset, dataset.id, IDLE_STATUS]);
    }

    private async collectDatasetInfo(dataset: Dataset) {
        await this.databaseDataLoaderActions.loadData(dataset.path, dataset.tableName);

        // create new dataset as one passed in args is readonly from the state.
        const newDataset: Dataset = {
            id: dataset.id,
            path: dataset.path,
            tableName: dataset.tableName,
            head: undefined,
        };

        await Promise.all([
            async () => {
                newDataset.profile = await this.databaseTableActions.getProfileColumns(dataset.tableName);
                newDataset.profile = newDataset.profile
                    .filter(row => row.name !== "duckdb_schema" && row.name !== "schema" && row.name !== "root");
            },
            async () => newDataset.sizeInBytes = await this.databaseDataLoaderActions.getDestinationSize(dataset.path),
            async () => newDataset.cardinality = await this.databaseTableActions.getCardinality(dataset.tableName),
            async () => newDataset.head = await this.databaseTableActions.getFirstN(dataset.tableName),
        ].map(asyncFunc => asyncFunc()));

        this.dataModelerStateManager.dispatch("addOrUpdateDatasetToState",
            [newDataset, false]);
    }
}
