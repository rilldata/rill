import type {DatabaseTableActions} from "../database/DatabaseTableActions";
import type {DatabaseColumnActions} from "../database/DatabaseColumnActions";
import type {DatabaseDataLoaderActions} from "../database/DatabaseDataLoaderActions";
import type {DataModelerStateManager} from "../state-actions/DataModelerStateManager";
import type {DataModelerActionAPI} from "$common/data-modeler-actions/DataModelerActionAPI";

export class DataModelerActions {
    protected dataModelerActionAPI: DataModelerActionAPI;

    constructor(protected readonly dataModelerStateManager: DataModelerStateManager,
                protected readonly databaseTableActions: DatabaseTableActions,
                protected readonly databaseColumnActions: DatabaseColumnActions,
                protected readonly databaseDataLoaderActions: DatabaseDataLoaderActions) {}

    public setDataModelerActionAPI(dataModelerActionAPI: DataModelerActionAPI): void {
        this.dataModelerActionAPI = dataModelerActionAPI;
    }
}
