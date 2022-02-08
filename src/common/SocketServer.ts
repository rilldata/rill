import { Server } from "socket.io";
import type {DataModelerService} from "$common/data-modeler-actions/DataModelerService";
import type {DataModelerStateService} from "$common/state-actions/DataModelerStateService";

export class SocketServer {
    private readonly server: Server;

    constructor(private readonly dataModelerService: DataModelerService,
                private readonly dataModelerStateService: DataModelerStateService,
                origin: string, private readonly port: number) {
        this.server = new Server({
            cors: { origin, methods: ["GET", "POST"] },
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

        this.server.listen(this.port);
    }

    public async destroy(): Promise<void> {
        this.server.close();
    }
}
