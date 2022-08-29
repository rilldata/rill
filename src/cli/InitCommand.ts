import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";
import { existsSync, mkdirSync, copyFileSync } from "fs";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export class InitCommand extends DataModelerCliCommand {
  private alreadyInitialised: boolean;

  public getCommand(): Command {
    return this.applyCommonSettings(
      new Command("init"),
      "Initialize a new project. The project location defaults to the current folder or you can use the --project option to specify a path. "
    )
      .option("--db <duckDbPath>", "Connect to an existing duckDB database. ")
      .option(
        "--copy",
        "Used with --db. Copy the duckDB database instead of directly modifying it. "
      )
      .action((opts, command) => {
        const { project } = command.optsWithGlobals();
        const projectPath = project ?? process.cwd();

        return this.createProjectAndRun(opts, projectPath);
      });
  }

  public createProjectAndRun(opts, projectPath: string) {
    InitCommand.makeDirectoryIfNotExists(projectPath);
    this.alreadyInitialised = existsSync(`${projectPath}/state`);

    if (
      !this.alreadyInitialised &&
      !InitCommand.verifyDuckDbPath(opts.db, opts.copy, projectPath)
    ) {
      console.log(`Failed to initialize a project under ${projectPath} `);
      return;
    }

    return this.run({
      projectPath,
      duckDbPath: opts.copy ? undefined : opts.db,
    });
  }

  protected async sendActions(): Promise<void> {
    if (!this.alreadyInitialised) {
      // add a single model.
      await this.dataModelerService.dispatch("addModel", [{}]);
      // set dummy asset as active to show onboarding steps
      await this.dataModelerService.dispatch("setActiveAsset", [
        EntityType.Model,
        undefined,
      ]);
      console.log(
        "\nYou have successfully initialized a new project with Rill Developer. "
      );
    } else {
      console.log(
        "\nA project has already been initialized in this directory. "
      );
    }
    console.log(
      "\nThis application is extremely alpha and we want to hear from you if you have any questions or ideas to share! " +
        "You can reach us in our Rill Discord server at https://bit.ly/3NSMKdT. "
    );
  }

  private static makeDirectoryIfNotExists(path: string) {
    if (!existsSync(path)) {
      console.log(`Directory ${path} doesn't exist. Creating the directory.`);
      // Use nodejs methods instead of running commands for making directory
      // This will ensure we can create the directory on all Operating Systems
      mkdirSync(path, { recursive: true });
    } else if (path !== process.cwd()) {
      console.log(
        `Directory ${path} already exist. Attempting to initialize the project. `
      );
    }
  }

  private static verifyDuckDbPath(
    duckDbPath: string,
    copy: boolean,
    projectPath: string
  ): boolean {
    if (!duckDbPath) return true;

    if (!existsSync(duckDbPath)) {
      console.log(
        `The duckDB database path provided ${duckDbPath} doesnt exist. `
      );
      return false;
    }

    console.log(
      `Importing tables from Duckdb database : ${duckDbPath} .\n` +
        `Please close any write connections to this database before connecting. `
    );

    if (copy) {
      copyFileSync(duckDbPath, `${projectPath}/stage.db`);
      console.log(
        "Copied over the database files. Any changes in Rill Developer won't be reflected in the other database file."
      );
    } else {
      console.log(
        `Note: Any source imports and drops will directly affect this connected database.`
      );
    }

    return true;
  }
}
