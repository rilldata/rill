import {DataModelerService, DataModelerActionsDefinition} from "$common/data-modeler-actions/DataModelerService";
import { io, Socket } from "socket.io-client";
import type {DataModelerStateService} from "$common/state-actions/DataModelerStateService";
import type {Patch} from "immer";

export class DataModelerSocketService extends DataModelerService {
    private socket: Socket;

    public constructor(dataModelerStateManager: DataModelerStateService, private readonly serverUrl: string) {
        super(dataModelerStateManager, null, []);
    }

    public async init(): Promise<void> {
        await super.init();
        this.socket = io(this.serverUrl);
        this.socket.on("patch", (patches: Array<Patch>) => this.dataModelerStateService.applyPatches(patches));
        this.socket.on("init-state", (initialState) => this.dataModelerStateService.updateState(initialState));
        // this.socket.on("connect", () => console.log("DataModelerActionSocketAPI Connected to server"));
        // this.socket.on("connect_error", (err) => console.log("connect_error", err));
        // this.socket.on("disconnect", (err) => console.log("disconnect", err));
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
