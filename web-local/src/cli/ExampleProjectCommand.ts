import { DataModelerCliCommand } from "./DataModelerCliCommand";
import { ImportTableCommand } from "./ImportTableCommand";
import { InitCommand } from "./InitCommand";
import { StartCommand } from "./StartCommand";
import { Command } from "commander";
import { execSync } from "node:child_process";
import Os from "os";
import { copyFileSync, existsSync, mkdirSync, readdirSync } from "fs";

function isWindows() {
  return Os.platform() === "win32";
}

export class ExampleProjectCommand extends DataModelerCliCommand {
  public getCommand(): Command {
    return this.applyCommonSettings(
      new Command("init-example"),
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
    console.log(`Initializing the example project ${project} ...`);
    await new InitCommand().createProjectAndRun({}, project);

    console.log("Downloading dataset for example project...");
    execSync(
      `curl -s http://pkg.rilldata.com/rill-developer-example/example-assets-0.6.zip ` +
        `--output "${project}/example-assets-0.6.zip"`,
      { stdio: "inherit" }
    );
    if (isWindows()) {
      execSync(
        `powershell -Command "Expand-Archive -Force ${project}/example-assets-0.6.zip ${project}/"`,

        {
          stdio: "inherit",
        }
      );
    } else {
      execSync(
        `unzip -o "${project}/example-assets-0.6.zip" ` + `-d "${project}/"`,
        {
          stdio: "inherit",
        }
      );
    }

    console.log("Importing example datasets into the project...");
    console.log("Ecommerce...");
    await new ImportTableCommand().run(
      {
        projectPath: project,
        profileWithUpdate: true,
      },
      `${project}/example-assets-0.6/data/ecomm-click-stream/e-shop-clothing.csv`,
      { force: true }
    );

    console.log("Global...");
    await new ImportTableCommand().run(
      {
        projectPath: project,
        profileWithUpdate: true,
      },
      `${project}/example-assets-0.6/data/global-landslide-catalog/global-landslide-catalog.csv`,
      { force: true }
    );

    console.log("Internet of Things...");
    await new ImportTableCommand().run(
      {
        projectPath: project,
        profileWithUpdate: true,
      },
      `${project}/example-assets-0.6/data/iot-env-sensor/iot-telemetry-data.csv`,
      { force: true }
    );

    console.log("Importing example SQL transformations into the project...");
    // this will handle escaping the folder
    readdirSync(`${project}/example-assets-0.6/models/`).forEach(
      (modelFile) =>
        modelFile.endsWith(".sql") &&
        copyFileSync(
          `${project}/example-assets-0.6/models/${modelFile}`,
          `${project}/models/${modelFile}`
        )
    );

    console.log("Cleaning up the project...");
    if (!existsSync(`${project}/data`)) {
      mkdirSync(`${project}/data`, {
        recursive: true,
      });
      execSync(`mv -v "${project}/example-assets-0.6/data" "${project}"`, {
        stdio: "inherit",
      });
    }

    execSync(`rm -rf "${project}/example-assets-0.6"`, {
      stdio: "inherit",
    });
    execSync(`rm -rf "${project}/example-assets-0.6.zip"`, {
      stdio: "inherit",
    });
    execSync(`rm -rf "${project}/__MACOSX"`, {
      stdio: "inherit",
    });

    console.log("Starting example...");
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
