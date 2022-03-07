import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";
import { execSync } from "node:child_process";

export class InitCommand extends DataModelerCliCommand {
    public getCommand(): Command {
        return new Command("init")
            .description("Initialize a new project either in the current folder or supplied folder.")
            .option("--project <projectPath>", "Optional path of project. Defaults to current directory.")
            .action(({ project }) => {
                return this.run({ projectPath: project });
            });
    }

    protected async sendActions(): Promise<void> {
        execSync(`mkdir -p ${this.projectPath}/models`);
        console.log("You have successfully initialized a new project with the Rill Data Modeler. " +
            "This application is extremely alpha and we want to hear from you if you have any questions or ideas to share! "+
            "You can reach us in our Rill Community Slack at https://bit.ly/3Mig8Jr.");
    }
}
