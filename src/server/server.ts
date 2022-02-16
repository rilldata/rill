import "../moduleAlias";
import {serverFactory} from "$common/serverFactory";
import {RootConfig} from "$common/config/RootConfig";

const {socketServer} = serverFactory(RootConfig.getDefaultConfig());
socketServer.init();
