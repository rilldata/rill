import { Server } from "socket.io";
import type {DataModelerActionAPI} from "$common/data-modeler-actions/DataModelerActionAPI";
import type {DataModelerStateManager} from "$common/state-actions/DataModelerStateManager";

export class SocketServer {
    private readonly server: Server;

    constructor(private readonly dataModelerActionAPI: DataModelerActionAPI,
                private readonly dataModelerStateManager: DataModelerStateManager,
                origin: string, private readonly port: number) {
        this.server = new Server({
            cors: { origin, methods: ["GET", "POST"] },
        });
    }

    public async init(): Promise<void> {
        await this.dataModelerActionAPI.init();

        this.dataModelerStateManager.subscribePatches((patches) => {
            this.server.emit("patch", patches);
        });

        this.server.on("connection", (socket) => {
            socket.emit("init-state", this.dataModelerStateManager.getCurrentState());
            socket.on("action", async (action, args) => {
                await this.dataModelerActionAPI.dispatch(action, args);
            });
        });

        this.server.listen(this.port);
    }

    public async destroy(): Promise<void> {
        this.server.close();
    }
}
