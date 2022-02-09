import { Server } from "socket.io";
import type {DataModelerService} from "$common/data-modeler-service/DataModelerService";
import type {DataModelerStateService} from "$common/data-modeler-state-service/DataModelerStateService";
import type {ServerConfig} from "$common/config/ServerConfig";

export class SocketServer {
    private readonly server: Server;

    constructor(private readonly dataModelerService: DataModelerService,
                private readonly dataModelerStateService: DataModelerStateService,
                private readonly serverConfig: ServerConfig) {
        this.server = new Server({
            cors: { origin: this.serverConfig.serverUrl, methods: ["GET", "POST"] },
        });
    }

    public async init(): Promise<void> {
        await this.dataModelerService.init();

        this.dataModelerStateService.subscribePatches((patches) => {
            this.server.emit("patch", patches);
        });

        this.server.on("connection", (socket) => {
            socket.emit("init-state", this.dataModelerStateService.getCurrentState());
            socket.on("action", async (action, args) => {
                await this.dataModelerService.dispatch(action, args);
            });
        });

        this.server.listen(this.serverConfig.socketPort);
    }

    public async destroy(): Promise<void> {
        this.server.close();
    }
}
