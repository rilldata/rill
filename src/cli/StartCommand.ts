import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import { Command } from "commander";

export class StartCommand extends DataModelerCliCommand {
    public getCommand(): Command {
        return new Command("start")
            .description("Starts the data-modeler UI.")
            .option("--project <projectPath>", "Optional path of project. Defaults to current directory.")
            .action(({project}) => {
                return this.run(project);
            });
    }

    protected sendActions(dataModelerService: DataModelerService, dataModelerStateService: DataModelerStateService): Promise<void> {
        // TODO
        return Promise.resolve(undefined);
    }
}
