import { Command } from "commander";
import { writeFileSync } from "fs";
import { execSync } from "node:child_process";
import { getCliInstances, SAVED_STATE_FILE } from "$cli/cliFactory";

export async function initProject(projectPath: string): Promise<void> {
    const computedProjectPath = projectPath ?? process.cwd();
    execSync(`mkdir -p ${computedProjectPath}`);
    const {dataModelerStateService} = await getCliInstances(computedProjectPath);
    writeFileSync(`${computedProjectPath}/${SAVED_STATE_FILE}`, JSON.stringify(dataModelerStateService.getCurrentState()));
    execSync(`mkdir -p ${computedProjectPath}/models`);
}

export function getInitCommand(): Command {
    return new Command("init")
        .description("Initialize a new project either in the current folder or supplied folder.")
        .argument("[path]", "Optional path to the project. Defaults to current directory.", process.cwd())
        .action((path) => {
            return initProject(path);
        });
}
