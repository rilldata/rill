import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";
import { existsSync, mkdirSync } from "fs";

export class InitCommand extends DataModelerCliCommand {
    private alreadyInitialised: boolean;

    public getCommand(): Command {
        return this.applyCommonSettings(
            new Command("init"),
            "Initialize a new project either in the current folder or supplied folder.",
        )
            .action((opts, command) => {
                const {project} = command.optsWithGlobals();

                const projectPath = project ?? process.cwd();
                InitCommand.makeDirectoryIfNotExists(projectPath);
                this.alreadyInitialised = existsSync(`${projectPath}/state`);

                return this.run({ projectPath });
            });
    }

    protected async sendActions(): Promise<void> {
        if (!existsSync(`${this.projectPath}/models`)) {
            mkdirSync(`${this.projectPath}/models`, {});
        }
        if (!this.alreadyInitialised) {
            // add a single model.
            this.dataModelerService.dispatch("addModel", [{}]);
            // make that model active.
            console.log("\nYou have successfully initialized a new project with the Rill Developer.");
        } else {
            console.log("\nA project in this directory has already been initialized.");
        }
        console.log("\nThis application is extremely alpha and we want to hear from you if you have any questions or ideas to share! "+
            "You can reach us in our Rill Community Slack at https://bit.ly/3Mig8Jr.");
    }

    private static makeDirectoryIfNotExists(path: string) {
        if (!existsSync(path)) {
            console.log(`Directory ${path} doest exist. Creating the directory.`);
            // Use nodejs methods instead of running commands for making directory
            // This will ensure we can create the directory on all Operating Systems
            mkdirSync(path, { recursive: true });
        } else if (path !== process.cwd()) {
            console.log(`Directory ${path} already exist. Attempting to init the project.`);
        }
    }
}
