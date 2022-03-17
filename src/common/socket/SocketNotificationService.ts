import type { NotificationService, Notification } from "$common/notifications/NotificationService";
import type { Server } from "socket.io";
import type { ClientToServerEvents, ServerToClientEvents } from "$common/socket/SocketInterfaces";

export class SocketNotificationService implements NotificationService {
    private server: Server<ClientToServerEvents, ServerToClientEvents>;

    public setSocketServer(server: Server<ClientToServerEvents, ServerToClientEvents>) {
        this.server = server;
    }

    public notify(notification: Notification): void {
        this.server?.emit("notification", notification);
    }
}
