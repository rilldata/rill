import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type { DatabaseService } from "$common/database-service/DatabaseService";
import type { DataModelerActions } from "$common/data-modeler-service/DataModelerActions";

export class RillDeveloperService {
  public constructor(
    protected readonly dataModelerStateService: DataModelerStateService,
    private readonly databaseService: DatabaseService,
    private readonly dataModelerActions: Array<DataModelerActions>
  ) {}
}
