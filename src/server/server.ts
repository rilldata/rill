import "../moduleAlias";
import {RootConfig} from "$common/config/RootConfig";
import { RillDeveloper } from "./RillDeveloper";
import { SocketServer } from "./SocketServer";
import type {SocketNotificationService} from "$common/socket/SocketNotificationService";

const config = new RootConfig({});
const rillDeveloper = RillDeveloper.getRillDeveloper(config);
const socketServer = new SocketServer(config, rillDeveloper.dataModelerService,
    rillDeveloper.dataModelerStateService, rillDeveloper.metricsService);
(rillDeveloper.notificationService as SocketNotificationService)
    .setSocketServer(socketServer.getSocketServer());
(async () => {
    await rillDeveloper.init();
    await socketServer.init();
    socketServer.getSocketServer().listen(config.server.socketPort);
})();
