import { EntityStateUpdatesHandler } from "./EntityStateUpdatesHandler";
import type { RootConfig } from "../../config/RootConfig";
import type {
  DataModelerActionsDefinition,
  DataModelerService,
} from "../../data-modeler-service/DataModelerService";
import type { DataProfileEntity } from "../entity-state-service/DataProfileEntity";
import { Throttler } from "../../utils/Throttler";

/**
 * Update handler that triggers action to profile if not already profiled.
 */
export class DerivedEntityUpdateHandler extends EntityStateUpdatesHandler<DataProfileEntity> {
  protected collectInfoThrottler = new Throttler();

  public constructor(
    protected readonly config: RootConfig,
    protected readonly dataModelerService: DataModelerService,
    protected readonly collectEntityInfoAction: keyof DataModelerActionsDefinition
  ) {
    super(config, dataModelerService);
  }

  public async handleEntityInit(entity: DataProfileEntity): Promise<void> {
    return this.handleModelProfiling(entity);
  }

  public async handleNewEntity(entity: DataProfileEntity): Promise<void> {
    return this.handleModelProfiling(entity);
  }

  public async handleUpdatedEntity(entity: DataProfileEntity): Promise<void> {
    return this.handleModelProfiling(entity);
  }

  private async handleModelProfiling(entity: DataProfileEntity): Promise<void> {
    // if the entity is already profiled or if profiling is disabled,
    // do not dispatch profiling action for this entity
    if (entity.profiled || !this.config.profileWithUpdate) return;
    // it is possible the collect info will take a long time.
    // this code might end up running multiple time by then.
    // add a throttler to make sure we don't call the collect info multiple times by then.
    this.collectInfoThrottler.throttle(
      entity.id,
      () => {
        // make sure to run it after a little delay
        // we need entry in both derived and persistent states
        // TODO: Find a better way to sync this
        setTimeout(async () => {
          try {
            await this.dataModelerService.dispatch(
              this.collectEntityInfoAction,
              [entity.id]
            );
          } catch (err) {
            // continue regardless of error
          }
          this.collectInfoThrottler.clear(entity.id);
        }, this.config.state.syncInterval);
      },
      5 * this.config.state.syncInterval
    );
  }
}
