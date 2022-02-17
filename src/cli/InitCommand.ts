import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";
import { execSync } from "node:child_process";

export class InitCommand extends DataModelerCliCommand {
    public getCommand(): Command {
        return new Command("init")
            .description("Initialize a new project either in the current folder or supplied folder.")
            .argument("[projectPath]", "Optional path to the project. Defaults to current directory.", process.cwd())
            .action((projectPath) => {
                return this.run(projectPath);
            });
    }

    protected async sendActions(): Promise<void> {
        execSync(`mkdir -p ${this.projectPath}/models`);
    }
}
