import { MetricsService } from "$common/metrics/MetricsService";
import type { MetricsActionDefinition } from "$common/metrics/MetricsService";
import type { Socket } from "socket.io-client";
import type { ClientToServerEvents, ServerToClientEvents } from "$common/socket/SocketInterfaces";
import type { RootConfig } from "$common/config/RootConfig";

export class MetricsSocketService extends MetricsService {
    private socket: Socket<ServerToClientEvents, ClientToServerEvents>;

    public constructor(config: RootConfig) {
        super(config, null, null, []);
    }

    public setSocket(socket: Socket<ServerToClientEvents, ClientToServerEvents>) {
        this.socket = socket;
    }

    public async dispatch<Action extends keyof MetricsActionDefinition>(
        action: Action, args: MetricsActionDefinition[Action],
    ): Promise<any> {
        this.socket?.emit("metrics", action, args);
    }
}
