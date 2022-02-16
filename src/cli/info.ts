import { Command } from "commander";
import { getCliInstances } from "$cli/cliFactory";
import type { DataModelerState } from "$lib/types";

export function displayProjectInfo(projectPath: string, state: DataModelerState) {
    console.log(`Project Path: ${projectPath}`);
    console.log("Tables:");
    state.tables.forEach(table => console.log(`${table.tableName} (${table.path})`));
    console.log("Models:");
    state.models.forEach(model => model.query && console.log(`${model.name}: ${model.query}`));
}

export async function showInfo(projectPath: string): Promise<void> {
    const computedProjectPath = projectPath ?? process.cwd();
    const {dataModelerStateService} = await getCliInstances(computedProjectPath);
    displayProjectInfo(computedProjectPath, dataModelerStateService.getCurrentState());
}

export function getInfoCommand(): Command {
    return new Command("info")
        .description("Displays info of a project.")
        .option("--project <projectPath>", "Optional path of project. Defaults to current directory.")
        .action(({project}) => {
            return showInfo(project);
        });
}
