import type {
  NotificationService,
  Notification,
} from "../notifications/NotificationService";
import type { Server } from "socket.io";
import type {
  ClientToServerEvents,
  ServerToClientEvents,
} from "./SocketInterfaces";

export class SocketNotificationService implements NotificationService {
  private server: Server<ClientToServerEvents, ServerToClientEvents>;

  public setSocketServer(
    server: Server<ClientToServerEvents, ServerToClientEvents>
  ) {
    this.server = server;
  }

  public notify(notification: Notification): void {
    this.server?.emit("notification", notification);
  }
}
