import {DataModelerActions} from ".//DataModelerActions";
import type {DataModelerState, Dataset} from "$lib/types";
import {newSource} from "$common/data-factory";
import {ColumnarItemType} from "$common/data-modeler-state-service/ProfileColumnStateActions";
import {IDLE_STATUS, RUNNING_STATUS} from "$common/constants";
import {sanitizeTableName} from "$lib/util/sanitize-table-name";

export class DatasetActions extends DataModelerActions {
    public async addOrUpdateDataset(currentState: DataModelerState, path: string): Promise<void> {
        const datasets = currentState.sources;
        const existingDataset = datasets.find(s => s.path === path);
        const dataset = {...(existingDataset || newSource())};
        dataset.path = path;
        dataset.tableName = sanitizeTableName(path);

        this.dataModelerStateService.dispatch("addOrUpdateDatasetToState",
            [dataset, !existingDataset]);
        this.dataModelerStateService.dispatch("setDatasetStatus",
            [ColumnarItemType.Dataset, dataset.id, RUNNING_STATUS]);

        try {
            await this.collectDatasetInfo(dataset);

            await this.dataModelerActionAPI.dispatch("collectProfileColumns",
                [dataset.id, ColumnarItemType.Dataset]);
        } catch (err) {
            console.log(err);
        }

        this.dataModelerStateService.dispatch("setDatasetStatus",
            [ColumnarItemType.Dataset, dataset.id, IDLE_STATUS]);
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
