import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";
import { ExpressServer } from "../server/ExpressServer";

export class StartCommand extends DataModelerCliCommand {
    public getCommand(): Command {
        return new Command("start")
            .description("Starts the data-modeler UI.")
            .option("--project <projectPath>", "Optional path of project. Defaults to current directory.")
            .action(({project}) => {
                return this.run({
                    projectPath: project, shouldInitState: false,
                    shouldSkipDatabase: false, profileWithUpdate: true,
                });
            });
    }

    protected sendActions(): Promise<void> {
        return new ExpressServer(this.config, this.dataModelerService, this.dataModelerStateService,
            this.notificationService).init();
    }
}
