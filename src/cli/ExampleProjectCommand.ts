import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";
import { execSync } from "node:child_process";
import { ImportTableCommand } from "$cli/ImportTableCommand";
import { InitCommand } from "$cli/InitCommand";
import { StartCommand } from "$cli/StartCommand";

export class ExampleProjectCommand extends DataModelerCliCommand {
  public getCommand(): Command {
    return this.applyCommonSettings(
      new Command("initialize-example-project"),
      "Initialize example project."
    ).action((opts, command: Command) => {
      let { project } = command.optsWithGlobals();
      if (!project) project = process.cwd() + "/rill-developer-example";

      return this.createExampleProject(project);
    });
  }

  protected async sendActions(): Promise<void> {
    // no-op
  }

  public async createExampleProject(project: string): Promise<void> {
    console.log(`Initializing the project example project ${project} ...`);
    await new InitCommand().createProjectAndRun({}, project);

    console.log("Downloading dataset for example project...");
    execSync(
      `curl -s http://pkg.rilldata.com/rill-developer-example/data/flightlist.zip ` +
        `--output ${project}/flightlist.zip`,
      { stdio: "inherit" }
    );
    execSync(`unzip ${project}/flightlist.zip ` + `-d ${project}/`, {
      stdio: "inherit",
    });

    console.log("Importing example dataset into the project...");
    await new ImportTableCommand().run(
      {
        projectPath: project,
        profileWithUpdate: true,
      },
      `${project}/data/flightlist_2022_02.csv`,
      {}
    );

    return new StartCommand().run({
      projectPath: project,
      shouldInitState: false,
      shouldSkipDatabase: false,
      profileWithUpdate: true,
    });
  }

  protected async teardown(): Promise<void> {
    // do not teardown as this will have a perpetual server
  }
}
