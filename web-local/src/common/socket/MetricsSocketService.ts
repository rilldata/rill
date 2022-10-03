import { MetricsService } from "../metrics-service/MetricsService";
import type { MetricsActionDefinition } from "../metrics-service/MetricsService";
import type { Socket } from "socket.io-client";
import type {
  ClientToServerEvents,
  ServerToClientEvents,
} from "./SocketInterfaces";
import type { RootConfig } from "../config/RootConfig";

export class MetricsSocketService extends MetricsService {
  private socket: Socket<ServerToClientEvents, ClientToServerEvents>;

  public constructor(config: RootConfig) {
    super(config, null, null, []);
  }

  public setSocket(socket: Socket<ServerToClientEvents, ClientToServerEvents>) {
    this.socket = socket;
  }

  public async dispatch<Action extends keyof MetricsActionDefinition>(
    action: Action,
    args: MetricsActionDefinition[Action]
  ): Promise<void> {
    this.socket?.emit("metrics", action, args);
  }
}
