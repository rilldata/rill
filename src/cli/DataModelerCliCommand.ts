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
    protected dataModelerStateSyncService: DataModelerStateSyncService;
    protected projectPath: string;
    protected config: RootConfig;

    private async init({ projectPath, shouldInitState, shouldSkipDatabase, profileWithUpdate }: CliRunArgs): Promise<void> {
        this.projectPath = projectPath ?? process.cwd();
        shouldInitState = shouldInitState ?? true;
        shouldSkipDatabase = shouldSkipDatabase ?? true;
        profileWithUpdate = profileWithUpdate ?? false;

        this.config = new RootConfig({
            database: new DatabaseConfig({ databaseName: DATABASE_NAME, skipDatabase: shouldSkipDatabase }),
            server: new ServerConfig({ serverPort: 8080, serveStaticFile: true }),
            projectFolder: this.projectPath, profileWithUpdate,
        });
        const {dataModelerService, dataModelerStateService, notificationService} = dataModelerServiceFactory(this.config);

        if (shouldInitState) {
            this.dataModelerStateSyncService = new DataModelerStateSyncService(
                this.config, dataModelerStateService.entityStateServices,
                dataModelerService, dataModelerStateService,
            );
            await this.dataModelerStateSyncService.init();
        }
        await dataModelerService.init();

        this.dataModelerService = dataModelerService;
        this.dataModelerStateService = dataModelerStateService;
        this.notificationService = notificationService;
    }

    private async teardown(): Promise<void> {
        await this.dataModelerStateService.destroy();
        await this.dataModelerStateSyncService?.destroy();
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
