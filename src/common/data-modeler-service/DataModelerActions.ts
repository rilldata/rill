import type {DataModelerStateService} from "../data-modeler-state-service/DataModelerStateService";
import type {DataModelerService} from "$common/data-modeler-service/DataModelerService";
import type {DatabaseService} from "$common/database-service/DatabaseService";

export class DataModelerActions {
    protected dataModelerService: DataModelerService;

    constructor(protected readonly dataModelerStateService: DataModelerStateService,
                protected readonly databaseService: DatabaseService) {}

    public setDataModelerActionAPI(dataModelerActionAPI: DataModelerService): void {
        this.dataModelerService = dataModelerActionAPI;
    }
}
