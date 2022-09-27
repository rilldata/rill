import type { RootConfig } from "../config/RootConfig";
import type { DataModelerService } from "../data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "../data-modeler-state-service/DataModelerStateService";

export abstract class DataConnection {
  protected constructor(
    protected readonly config: RootConfig,
    protected readonly dataModelerService: DataModelerService,
    protected readonly dataModelerStateService: DataModelerStateService
  ) {}

  public abstract init(): Promise<void>;
  public abstract sync(): Promise<void>;
  public abstract destroy(): Promise<void>;
}
