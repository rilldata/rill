import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
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

    protected sendActions(tableName: string): Promise<void> {
        // TODO
        return Promise.resolve(undefined);
    }
}
