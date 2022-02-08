import type {DataModelerStateService} from "../state-actions/DataModelerStateService";
import type {DataModelerService} from "$common/data-modeler-actions/DataModelerService";
import type {DatabaseService} from "$common/database/DatabaseService";

export class DataModelerActions {
    protected dataModelerActionAPI: DataModelerService;

    constructor(protected readonly dataModelerStateService: DataModelerStateService,
                protected readonly databaseService: DatabaseService) {}

    public setDataModelerActionAPI(dataModelerActionAPI: DataModelerService): void {
        this.dataModelerActionAPI = dataModelerActionAPI;
    }
}
