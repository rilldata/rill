import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type { DataModelerState } from "$lib/types";
import { existsSync, readFileSync, writeFileSync } from "fs";
import { dataModelerServiceFactory } from "$common/serverFactory";
import { RootConfig } from "$common/config/RootConfig";
import { DatabaseConfig } from "$common/config/DatabaseConfig";
import type { Command } from "commander";

const SAVED_STATE_FILE = "saved-state.json";
const DATABASE_NAME = "stage.db";

export abstract class DataModelerCliCommand {
    private dataModelerService: DataModelerService;
    private dataModelerStateService: DataModelerStateService;
    private projectPath: string;

    private async init(projectPath: string): Promise<void> {
        this.projectPath = projectPath ?? process.cwd();

        let initialState: DataModelerState;
        if (existsSync(`${this.projectPath}/${SAVED_STATE_FILE}`)) {
            initialState = JSON.parse(readFileSync(`${this.projectPath}/${SAVED_STATE_FILE}`).toString());
        }

        const {dataModelerService, dataModelerStateService} = dataModelerServiceFactory(new RootConfig({
            database: new DatabaseConfig({ databaseName: `${this.projectPath}/${DATABASE_NAME}` }),
        }));
        await dataModelerService.init(initialState);

        this.dataModelerService = dataModelerService;
        this.dataModelerStateService = dataModelerStateService;
    }

    private async teardown(): Promise<void> {
        writeFileSync(`${this.projectPath}/${SAVED_STATE_FILE}`,
            JSON.stringify(this.dataModelerStateService.getCurrentState()));
    }

    protected async run(projectPath: string, ...args: Array<any>): Promise<void> {
        await this.init(projectPath);
        await this.sendActions(this.dataModelerService, this.dataModelerStateService,
            projectPath, ...args);
        await this.teardown();
    }

    protected abstract sendActions(dataModelerService: DataModelerService,
                                   dataModelerStateService: DataModelerStateService,
                                   projectPath: string, ...args: Array<any>): Promise<void>;
    
    public abstract getCommand(): Command;
}
