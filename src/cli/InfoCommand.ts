import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import { Command } from "commander";
import type { DataModelerState } from "$lib/types";

export class InfoCommand extends DataModelerCliCommand {
    public getCommand(): Command {
        return new Command("info")
            .description("Displays info of a project.")
            .option("--project <projectPath>", "Optional path of project. Defaults to current directory.")
            .action(({project}) => {
                return this.run(project);
            });
    }

    protected async sendActions(dataModelerService: DataModelerService, dataModelerStateService: DataModelerStateService,
                          projectPath: string): Promise<void> {
        InfoCommand.displayProjectInfo(projectPath, dataModelerStateService.getCurrentState());
    }

    public static displayProjectInfo(projectPath: string, state: DataModelerState) {
        console.log(`Project Path: ${projectPath}`);
        console.log("Tables:");
        state.tables.forEach(table => console.log(`${table.tableName} (${table.path})`));
        console.log("Models:");
        state.models.forEach(model => model.query && console.log(`${model.name}: ${model.query}`));
    }
}
