import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import { dataModelerServiceFactory } from "$common/serverFactory";
import { RootConfig } from "$common/config/RootConfig";
import { DatabaseConfig } from "$common/config/DatabaseConfig";
import type { Command } from "commander";
import type { SocketNotificationService } from "$common/socket/SocketNotificationService";
import { StateConfig } from "$common/config/StateConfig";
import { ServerConfig } from "$common/config/ServerConfig";
import { execSync } from "node:child_process";
import {
    DataModelerStateSyncService
} from "$common/data-modeler-state-service/sync-service/DataModelerStateSyncService";

const SAVED_STATE_FILE = "saved-state.json";
const DATABASE_NAME = "stage.db";

export interface CliRunArgs {
    projectPath: string;
    shouldInitState?: boolean;
    shouldSkipDatabase?: boolean;
}

export abstract class DataModelerCliCommand {
    protected dataModelerService: DataModelerService;
    protected dataModelerStateService: DataModelerStateService;
    protected notificationService: SocketNotificationService;
    protected dataModelerStateSyncService: DataModelerStateSyncService;
    protected projectPath: string;
    protected config: RootConfig;

    private async init({ projectPath, shouldInitState, shouldSkipDatabase }: CliRunArgs): Promise<void> {
        this.projectPath = projectPath ?? process.cwd();
        shouldInitState = shouldInitState ?? true;
        shouldSkipDatabase = shouldSkipDatabase ?? true;

        this.config = new RootConfig({
            database: new DatabaseConfig({ databaseName: `${this.projectPath}/${DATABASE_NAME}`, skipDatabase: shouldSkipDatabase }),
            state: new StateConfig({ savedStateFile: `${this.projectPath}/${SAVED_STATE_FILE}` }),
            server: new ServerConfig({ serverPort: 8080, serveStaticFile: true }),
            projectFolder: this.projectPath, profileWithUpdate: false,
        });
        const {dataModelerService, dataModelerStateService, notificationService} = dataModelerServiceFactory(this.config);
        execSync(`mkdir -p ${this.projectPath}`);

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
    
    public abstract getCommand(): Command;
}
