import { DataModelerService } from "../data-modeler-service/DataModelerService";
import type { DataModelerActionsDefinition } from "../data-modeler-service/DataModelerService";
import { io } from "socket.io-client";
import type { Socket } from "socket.io-client";
import type { DataModelerStateService } from "../data-modeler-state-service/DataModelerStateService";
import type { ServerConfig } from "../config/ServerConfig";
import type {
  ClientToServerEvents,
  ServerToClientEvents,
} from "./SocketInterfaces";
import type { ActionResponse } from "../data-modeler-service/response/ActionResponse";
import type { MetricsActionDefinition } from "../metrics-service/MetricsService";

/**
 * {@link DataModelerService} implementation that sits on the client side.
 * Forwards dispatched actions to the socket server.
 * Also listens to immer patches from the socket server and applies to the DataModelerStateService.
 */
export class DataModelerSocketService extends DataModelerService {
  private socket: Socket<ServerToClientEvents, ClientToServerEvents>;

  public constructor(
    dataModelerStateManager: DataModelerStateService,
    private readonly serverConfig: ServerConfig
  ) {
    super(dataModelerStateManager, null, null, null, []);
  }

  public getSocket(): Socket<ServerToClientEvents, ClientToServerEvents> {
    return this.socket;
  }

  public async init(): Promise<void> {
    await super.init();
    this.socket = io(this.serverConfig.socketUrl);

    this.socket.on("patch", (entityType, stateType, patches) =>
      this.dataModelerStateService.applyPatches(entityType, stateType, patches)
    );
    this.socket.on("initialState", (initialState) =>
      this.dataModelerStateService.updateState(initialState)
    );
  }

  public async dispatch<Action extends keyof DataModelerActionsDefinition>(
    action: Action,
    args: DataModelerActionsDefinition[Action]
  ): Promise<ActionResponse> {
    return new Promise((resolve) =>
      this.socket.emit("action", action, args, resolve)
    );
  }

  public async fireEvent<Event extends keyof MetricsActionDefinition>(
    event: Event,
    args: MetricsActionDefinition[Event]
  ): Promise<void> {
    this.socket.emit("metrics", event, args);
  }

  public async destroy(): Promise<void> {
    this.socket.disconnect();
  }
}
