import {DataModelerService} from "$common/data-modeler-service/DataModelerService";
import type {DataModelerActionsDefinition} from "$common/data-modeler-service/DataModelerService";
import { io } from "socket.io-client";
import type { Socket } from "socket.io-client";
import type {DataModelerStateService} from "$common/data-modeler-state-service/DataModelerStateService";
import type {ServerConfig} from "$common/config/ServerConfig";
import type { ClientToServerEvents, ServerToClientEvents } from "$common/socket/SocketInterfaces";

export class DataModelerSocketService extends DataModelerService {
    private socket: Socket<ServerToClientEvents, ClientToServerEvents>;

    public constructor(dataModelerStateManager: DataModelerStateService,
                       private readonly serverConfig: ServerConfig) {
        super(dataModelerStateManager, null, null, []);
    }

    public getSocket(): Socket<ServerToClientEvents, ClientToServerEvents> {
        return this.socket;
    }

    public async init(): Promise<void> {
        await super.init();
        this.socket = io(this.serverConfig.socketUrl);
        this.socket.on("patch", (patches) => this.dataModelerStateService.applyPatches(patches));
        this.socket.on("initialState", (initialState) => this.dataModelerStateService.updateState(initialState));
    }

    public async dispatch<Action extends keyof DataModelerActionsDefinition>(
        action: Action, args: DataModelerActionsDefinition[Action],
    ): Promise<void> {
        this.socket.emit("action", action, args);
    }

    public async destroy(): Promise<void> {
        this.socket.disconnect();
    }
}
