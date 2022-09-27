import { RillActionsChannel } from "../utils/RillActionsChannel";
import type {
  EntityType,
  StateType,
} from "../data-modeler-state-service/entity-state-service/EntityStateService";
import type {
  EntityRecordMapType,
  EntityStateServicesMapType,
} from "../data-modeler-state-service/entity-state-service/EntityStateServicesMap";

/**
 * The class that will contain context for a request like user etc.
 * For now, it will have the {@link RillActionsChannel} instance mainly along with target entity details.
 */
export class RillRequestContext<
  ET extends EntityType = EntityType,
  ST extends StateType = StateType
> {
  public entityStateService: EntityStateServicesMapType[ET][ST];

  /**
   * ID of the primary target entity
   */
  public id: string;
  public entityType: ET;
  public stateType: ST;
  public record: EntityRecordMapType[ET][ST];

  public constructor(
    public readonly actionsChannel: RillActionsChannel,
    public readonly level = 0
  ) {}

  public setEntityStateService(
    entityStateService: EntityStateServicesMapType[ET][ST]
  ) {
    this.entityStateService = entityStateService;
  }

  public setEntityInfo(id: string, entityType: ET, stateType: ST) {
    this.id = id;
    this.entityType = entityType;
    this.stateType = stateType;
    this.record = this.entityStateService.getById(
      id
    ) as EntityRecordMapType[ET][ST];
  }

  public static getNewContext() {
    return new RillRequestContext(new RillActionsChannel());
  }
}
