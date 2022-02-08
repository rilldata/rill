import {DataModelerActionAPI, DataModelerActionsDefinition} from "$common/data-modeler-actions/DataModelerActionAPI";
import { io, Socket } from "socket.io-client";
import type {DataModelerStateManager} from "$common/state-actions/DataModelerStateManager";
import type {Patch} from "immer";

export class DataModelerActionSocketAPI extends DataModelerActionAPI {
    private socket: Socket;

    public constructor(dataModelerStateManager: DataModelerStateManager, private readonly serverUrl: string) {
        super(dataModelerStateManager, null, []);
    }

    public async init(): Promise<void> {
        await super.init();
        this.socket = io(this.serverUrl);
        this.socket.on("patch", (patches: Array<Patch>) => this.dataModelerStateManager.applyPatches(patches));
        this.socket.on("init-state", (initialState) => this.dataModelerStateManager.updateState(initialState));
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
