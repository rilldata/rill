import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";

export class DropTableCommand extends DataModelerCliCommand {
    public getCommand(): Command {
        return this.applyCommonSettings(
            new Command("drop-table"),
            "Drop a table.",
        )
            .argument("<tableName>", "Name of the table to drop.")
            .action((tableName, opts, command: Command) => {
                const {project} = command.optsWithGlobals();
                return this.run({ projectPath: project }, tableName);
            });
    }

    protected sendActions(tableName: string): Promise<void> {
        // TODO
        return Promise.resolve(undefined);
    }
}
