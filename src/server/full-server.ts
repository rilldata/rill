import "../moduleAlias";
import {RootConfig} from "$common/config/RootConfig";
import {RillDeveloper} from "$common/RillDeveloper";
import {ExpressServer} from "./ExpressServer";
import type {SocketNotificationService} from "$common/socket/SocketNotificationService";
import {ServerConfig} from "$common/config/ServerConfig";

/**
 * Use this script when developing only backend.
 * Not all features are available right now when developing both backend and frontend.
 */
const config = new RootConfig({
    server: new ServerConfig({ serverPort: 8080, serveStaticFile: true })
});
const rillDeveloper = RillDeveloper.getRillDeveloper(config);
const expressServer = new ExpressServer(
    config, rillDeveloper.dataModelerService,
    rillDeveloper.dataModelerStateService,
    rillDeveloper.notificationService as SocketNotificationService,
    rillDeveloper.metricsService,
);
(async () => {
    await rillDeveloper.init();
    await expressServer.init();
})();
