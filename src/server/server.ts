import "../moduleAlias";
import {serverFactory} from "$common/serverFactory";
import {RootConfig} from "$common/config/RootConfig";

const config = new RootConfig({});
const {socketServer} = serverFactory(config);
(async () => {
    await socketServer.init();
    socketServer.getSocketServer().listen(config.server.socketPort);
})();
