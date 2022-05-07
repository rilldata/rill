import express from "express";
import http from "http";
import type { RootConfig } from "$common/config/RootConfig";
import { SocketServer } from "$server/SocketServer";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type { SocketNotificationService } from "$common/socket/SocketNotificationService";
import type { MetricsService } from "$common/metrics-service/MetricsService";

const STATIC_FILES = `${__dirname}/../../build`;

export class ExpressServer {
    private readonly app: express.Application;
    private readonly server: http.Server;
    private readonly socketServer: SocketServer;

    constructor(private readonly config: RootConfig,
                private readonly dataModelerService: DataModelerService,
                dataModelerStateService: DataModelerStateService,
                notificationService: SocketNotificationService,
                metricsService: MetricsService) {
        this.app = express();
        this.server = http.createServer(this.app);

        this.socketServer = new SocketServer(config, dataModelerService,
            dataModelerStateService, metricsService, this.server);
        notificationService.setSocketServer(this.socketServer.getSocketServer());

        if (config.server.serveStaticFile) {
            this.app.use(express.static(STATIC_FILES));
        }
    }

    public async init(): Promise<void> {
        await this.socketServer.init();
        this.server.listen(this.config.server.serverPort);
        console.log(`Server started at ${this.config.server.serverUrl}`);
    }
}
