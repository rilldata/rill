import {RootConfig} from "$common/config/RootConfig";
import {RillDeveloper} from "$common/RillDeveloper";
import {SocketServer} from "$common/socket/SocketServer";

/**
 * Sveltekit hook to start the dev UI + server at the same time
 */

const config = new RootConfig({});
const rillDeveloper = RillDeveloper.getRillDeveloper(config);
const socketServer =  new SocketServer(config, rillDeveloper.dataModelerService,
    rillDeveloper.dataModelerStateService, rillDeveloper.metricsService);
let socketStarted = false;

async function startSocket() {
    if (socketStarted) return;
    socketStarted = true;

    await rillDeveloper.init();
    await socketServer.init();
    socketServer.getSocketServer().listen(config.server.socketPort);
}

/** @type {import('@sveltejs/kit').Handle} */
export async function handle({ resolve, request }) {
    await startSocket()
    return resolve(request);
}
