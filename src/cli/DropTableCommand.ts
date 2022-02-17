import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import { Command } from "commander";

export class DropTableCommand extends DataModelerCliCommand {
    public getCommand(): Command {
        return new Command("drop-table")
            .description("Drops a table.")
            .argument("<tableName>", "Name of the table to drop.")
            .option("--project <projectPath>", "Optional path of project. Defaults to current directory.")
            .action((tableName, {project}) => {
                return this.run(project, tableName);
            });
    }

    protected sendActions(dataModelerService: DataModelerService, dataModelerStateService: DataModelerStateService,
                          projectPath: string, tableName: string): Promise<void> {
        // TODO
        return Promise.resolve(undefined);
    }
}
