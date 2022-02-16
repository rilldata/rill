import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import { Command } from "commander";

export class DropTableCommand extends DataModelerCliCommand {
    public getCommand(): Command {
        return new Command("drop-table")
            .description("Drops a table.")
            .argument("<tableName>", "Name of the table to drop.")
            .action((tableName) => {
                // TODO
                console.log(tableName);
            });
    }

    protected sendActions(dataModelerService: DataModelerService, dataModelerStateService: DataModelerStateService,
                          projectPath: string, ...args: Array<any>): Promise<void> {
        // TODO
        return Promise.resolve(undefined);
    }
}
