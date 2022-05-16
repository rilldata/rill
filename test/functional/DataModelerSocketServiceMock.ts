import {
  DataModelerActionsDefinition,
  DataModelerService,
} from "$common/data-modeler-service/DataModelerService";
import type {
  DataModelerStateService,
  EntityTypeAndStates,
} from "$common/data-modeler-state-service/DataModelerStateService";
import type { SocketServerMock } from "./SocketServerMock";
import type { Patch } from "immer";
import type {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export class DataModelerSocketServiceMock extends DataModelerService {
  public socketServerMock: SocketServerMock;

  public constructor(dataModelerStateService: DataModelerStateService) {
    super(dataModelerStateService, null, null, null, []);
  }

  public async init(): Promise<void> {
    await this.dataModelerStateService.init();
  }

  public initialState(states: EntityTypeAndStates) {
    this.dataModelerStateService.updateState(states);
  }

  public applyPatches(
    entityType: EntityType,
    stateType: StateType,
    patches: Array<Patch>
  ) {
    this.dataModelerStateService.applyPatches(entityType, stateType, patches);
  }

  public async dispatch<Action extends keyof DataModelerActionsDefinition>(
    action: Action,
    args: DataModelerActionsDefinition[Action]
  ): Promise<any> {
    return this.socketServerMock.dispatch(action, args);
  }
}
