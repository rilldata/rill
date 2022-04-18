import "../moduleAlias";
import {RootConfig} from "$common/config/RootConfig";
import {RillDeveloper} from "$common/RillDeveloper";
import {SocketServer} from "$common/socket/SocketServer";

const config = new RootConfig({});
const rillDeveloper = RillDeveloper.getRillDeveloper(config);
const socketServer =  new SocketServer(config, rillDeveloper.dataModelerService,
    rillDeveloper.dataModelerStateService, rillDeveloper.metricsService);
(async () => {
    await rillDeveloper.init();
    await socketServer.init();
    socketServer.getSocketServer().listen(config.server.socketPort);
})();
