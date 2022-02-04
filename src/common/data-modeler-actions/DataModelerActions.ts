import type {DuckDBTableAPI} from "../duckdb/DuckDBTableAPI";
import type {DuckDBColumnAPI} from "../duckdb/DuckDBColumnAPI";
import type {DuckDBDataLoaderAPI} from "../duckdb/DuckDBDataLoaderAPI";
import type {DataModelerStateManager} from "../state-actions/DataModelerStateManager";
import type {DataModelerActionAPI} from "$common/data-modeler-actions/DataModelerActionAPI";

export class DataModelerActions {
    protected dataModelerActionAPI: DataModelerActionAPI;

    constructor(protected readonly dataModelerStateManager: DataModelerStateManager,
                protected readonly duckDBTableAPI: DuckDBTableAPI,
                protected readonly duckDBColumnAPI: DuckDBColumnAPI,
                protected readonly duckDBDataLoaderAPI: DuckDBDataLoaderAPI) {}

    public setDataModelerActionAPI(dataModelerActionAPI: DataModelerActionAPI): void {
        this.dataModelerActionAPI = dataModelerActionAPI;
    }
}
