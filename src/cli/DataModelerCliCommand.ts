import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import { dataModelerServiceFactory } from "$common/serverFactory";
import { RootConfig } from "$common/config/RootConfig";
import { DatabaseConfig } from "$common/config/DatabaseConfig";
import type { Command } from "commander";
import type { SocketNotificationService } from "$common/socket/SocketNotificationService";
import { ServerConfig } from "$common/config/ServerConfig";
import {
    DataModelerStateSyncService
} from "$common/data-modeler-state-service/sync-service/DataModelerStateSyncService";
import { clientFactory } from "$common/clientFactory";
import { isPortOpen } from "$common/utils/isPortOpen";
import type { MetricsService } from "$common/metrics/MetricsService";

const DATABASE_NAME = "stage.db";

export interface CliRunArgs {
    projectPath: string;
    shouldInitState?: boolean;
    shouldSkipDatabase?: boolean;
    profileWithUpdate?: boolean;
}

export abstract class DataModelerCliCommand {
    protected dataModelerService: DataModelerService;
    protected dataModelerStateService: DataModelerStateService;
    protected notificationService: SocketNotificationService;
    protected metricsService: MetricsService;
    protected dataModelerStateSyncService: DataModelerStateSyncService;
    protected projectPath: string;
    protected config: RootConfig;
    protected isClient: boolean;

    private async init(cliRunArgs: CliRunArgs): Promise<void> {
        this.projectPath = cliRunArgs.projectPath ?? process.cwd();
        cliRunArgs.shouldInitState ??= true;
        cliRunArgs.shouldSkipDatabase ??= true;
        cliRunArgs.profileWithUpdate ??= false;

        this.config = new RootConfig({
            database: new DatabaseConfig({ databaseName: DATABASE_NAME, skipDatabase: cliRunArgs.shouldSkipDatabase }),
            server: new ServerConfig({ serverPort: 8080, serveStaticFile: true }),
            projectFolder: this.projectPath, profileWithUpdate: cliRunArgs.profileWithUpdate,
        });

        const isServerRunning = await isPortOpen(this.config.server.socketPort);

        if (isServerRunning) {
            await this.initClientInstances();
        } else {
            // database should be started when server is not running.
            // We can write to database in this case
            this.config.database.skipDatabase = false;
            await this.initServerInstances(cliRunArgs);
        }

        await this.dataModelerStateSyncService?.init();
        await this.dataModelerService.init();

        this.isClient = isServerRunning;
    }

    private async teardown(): Promise<void> {
        await this.dataModelerStateSyncService?.destroy();
        await this.dataModelerService.destroy();
    }

    private async initServerInstances({ shouldInitState }: CliRunArgs) {
        const {dataModelerService, dataModelerStateService,
            notificationService, metricsService} = dataModelerServiceFactory(this.config);

        if (shouldInitState) {
            this.dataModelerStateSyncService = new DataModelerStateSyncService(
                this.config, dataModelerStateService.entityStateServices,
                dataModelerService, dataModelerStateService,
            );
        }

        this.dataModelerService = dataModelerService;
        this.dataModelerStateService = dataModelerStateService;
        this.notificationService = notificationService;
        this.metricsService = metricsService;
    }

    private async initClientInstances() {
        const {dataModelerService, dataModelerStateService, metricsService} = clientFactory(this.config);
        this.dataModelerService = dataModelerService;
        this.dataModelerStateService = dataModelerStateService;
        this.metricsService = metricsService;
    }

    protected async run(cliRunArgs: CliRunArgs, ...args: Array<any>): Promise<void> {
        await this.init(cliRunArgs);
        await this.sendActions(...args);
        await this.teardown();
    }

    protected abstract sendActions(...args: Array<any>): Promise<void>;

    protected applyCommonSettings(command: Command, description: string): Command {
        return command
            .description(description)
            // override default help text to add capital D for display
            .helpOption("-h, --help", "Display help for command.")
            // common across all commands
            .option("--project <projectPath>", "Optional path of project. Defaults to current directory.");
    }

    public abstract getCommand(): Command;
}
