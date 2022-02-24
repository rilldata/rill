import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import { dataModelerServiceFactory } from "$common/serverFactory";
import { RootConfig } from "$common/config/RootConfig";
import { DatabaseConfig } from "$common/config/DatabaseConfig";
import type { Command } from "commander";
import type { SocketNotificationService } from "$common/socket/SocketNotificationService";
import { StateConfig } from "$common/config/StateConfig";
import { ServerConfig } from "$common/config/ServerConfig";

const SAVED_STATE_FILE = "saved-state.json";
const DATABASE_NAME = "stage.db";

export abstract class DataModelerCliCommand {
    protected dataModelerService: DataModelerService;
    protected dataModelerStateService: DataModelerStateService;
    protected notificationService: SocketNotificationService;
    protected projectPath: string;
    protected config: RootConfig;

    private async init(projectPath: string): Promise<void> {
        this.projectPath = projectPath ?? process.cwd();

        this.config = new RootConfig({
            database: new DatabaseConfig({ databaseName: `${this.projectPath}/${DATABASE_NAME}` }),
            state: new StateConfig({ savedStateFile: `${this.projectPath}/${SAVED_STATE_FILE}` }),
            server: new ServerConfig({ serverPort: 8080, serveStaticFile: true }),
        });
        const {dataModelerService, dataModelerStateService, notificationService} = dataModelerServiceFactory(this.config);
        await dataModelerService.init();

        this.dataModelerService = dataModelerService;
        this.dataModelerStateService = dataModelerStateService;
        this.notificationService = notificationService;
    }

    private async teardown(): Promise<void> {
        await this.dataModelerStateService.destroy();
    }

    protected async run(projectPath: string, ...args: Array<any>): Promise<void> {
        await this.init(projectPath);
        await this.sendActions(...args);
        await this.teardown();
    }

    protected abstract sendActions(...args: Array<any>): Promise<void>;
    
    public abstract getCommand(): Command;
}
