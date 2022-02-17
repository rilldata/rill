import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";
import { ExpressServer } from "../server/ExpressServer";

export class StartCommand extends DataModelerCliCommand {
    public getCommand(): Command {
        return new Command("start")
            .description("Starts the data-modeler UI.")
            .option("--project <projectPath>", "Optional path of project. Defaults to current directory.")
            .action(({project}) => {
                return this.run(project);
            });
    }

    protected sendActions(): Promise<void> {
        return new ExpressServer(this.dataModelerService, this.dataModelerStateService,
            this.notificationService, this.config).init();
    }
}
