import {DataModelerActions} from "../data-modeler-actions/DataModelerActions";
import type {DataModelerState, Dataset} from "$lib/types";
import {newSource} from "$common/data-factory";
import {ColumnarItemType} from "$common/state-actions/ProfileColumnStateActions";

export class DatasetActions extends DataModelerActions {
    public async addOrUpdateDataset(currentState: DataModelerState, path: string): Promise<void> {
        const datasets = currentState.sources;
        const existingDataset = datasets.find(s => s.path === path);
        const dataset = {...(existingDataset || newSource())};
        dataset.path = path;
        dataset.name = path.split("/").slice(-1)[0].replace(/[.-]/g, "_");

        try {
            await this.collectDatasetInfo(dataset);
            this.dataModelerStateManager.dispatch("addOrUpdateDatasetToState",
                [dataset, !!existingDataset]);

            await this.dataModelerActionAPI.dispatch("collectProfileColumns",
                [dataset.id, ColumnarItemType.Dataset]);
        } catch (err) {
            console.log(err);
        }
    }

    private async collectDatasetInfo(dataset: Dataset) {
        if (!("profile" in dataset && dataset.profile.length)) {
            await this.duckDBDataLoaderAPI.loadData(dataset.path, dataset.name);
            dataset.profile = await this.duckDBTableAPI.getProfileColumns(dataset.name);
            dataset.profile = dataset.profile.filter(row => row.name !== "duckdb_schema" && row.name !== "schema" && row.name !== "root");
        }
        dataset.sizeInBytes = await this.duckDBTableAPI.getDestinationSize(dataset.path);
        dataset.cardinality = await this.duckDBTableAPI.getCardinality(dataset.name);
        dataset.head = await this.duckDBTableAPI.getFirstN(dataset.name);
    }
}
