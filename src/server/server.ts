import "../moduleAlias";
import {serverFactory} from "$common/serverFactory";
import {RootConfig} from "$common/config/RootConfig";

const config = new RootConfig({});
const {dataModelerService, socketServer} = serverFactory(config);
(async () => {
    await dataModelerService.init();
    dataModelerService.dispatch("updateTablesFromSource", [config.database.parquetFolder]);
    await socketServer.init();
    socketServer.getSocketServer().listen(config.server.socketPort);
})();
