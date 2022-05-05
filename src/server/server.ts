import "../moduleAlias";
import {RootConfig} from "$common/config/RootConfig";
import {RillDeveloper} from "$common/RillDeveloper";
import type {SocketNotificationService} from "$common/socket/SocketNotificationService";
import {ExpressServer} from "./ExpressServer";
import {ServerConfig} from "$common/config/ServerConfig";

const config = new RootConfig({
    server: new ServerConfig({ serveStaticFile: true })
});
const rillDeveloper = RillDeveloper.getRillDeveloper(config);
const expressServer = new ExpressServer(config,
    rillDeveloper.dataModelerService, rillDeveloper.dataModelerStateService,
    rillDeveloper.notificationService as SocketNotificationService, rillDeveloper.metricsService);
(async () => {
    await rillDeveloper.init();
    await expressServer.init();
})();
