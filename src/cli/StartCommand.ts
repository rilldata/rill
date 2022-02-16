import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import { Command } from "commander";

export class StartCommand extends DataModelerCliCommand {
    public getCommand(): Command {
        return new Command("start")
            .description("Starts the data-modeler UI.")
            .action(() => {
                // TODO
            });
    }

    protected sendActions(dataModelerService: DataModelerService, dataModelerStateService: DataModelerStateService): Promise<void> {
        return Promise.resolve(undefined);
    }
}
