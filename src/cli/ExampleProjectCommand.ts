import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { ImportTableCommand } from "$cli/ImportTableCommand";
import { InitCommand } from "$cli/InitCommand";
import { StartCommand } from "$cli/StartCommand";
import { Command } from "commander";
import { execSync } from "node:child_process";

export class ExampleProjectCommand extends DataModelerCliCommand {
  public getCommand(): Command {
    return this.applyCommonSettings(
      new Command("init-example-project"),
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
      `curl -s http://pkg.rilldata.com/rill-developer-example/example-assets.zip ` +
        `--output ${project}/example-assets.zip`,
      { stdio: "inherit" }
    );
    execSync(`unzip ${project}/example-assets.zip ` + `-d ${project}/`, {
      stdio: "inherit",
    });

    console.log("Importing example datasets into the project...");
    console.log("Adtech...");
    await new ImportTableCommand().run(
      {
        projectPath: project,
        profileWithUpdate: true,
      },
      `${project}/example-assets/data/adtech-ad-click/adtech-item-data.csv`,
      {}
    );
    await new ImportTableCommand().run(
      {
        projectPath: project,
        profileWithUpdate: true,
      },
      `${project}/example-assets/data/adtech-ad-click/adtech-train.csv`,
      {}
    );
    await new ImportTableCommand().run(
      {
        projectPath: project,
        profileWithUpdate: true,
      },
      `${project}/example-assets/data/adtech-ad-click/adtech-view-log.csv`,
      {}
    );

    console.log("Crypto...");
    await new ImportTableCommand().run(
      {
        projectPath: project,
        profileWithUpdate: true,
      },
      `${project}/example-assets/data/crypto-bitcoin/crypto-bitstamp-usd.csv`,
      {}
    );

    console.log("Ecommerce...");
    await new ImportTableCommand().run(
      {
        projectPath: project,
        profileWithUpdate: true,
      },
      `${project}/example-assets/data/ecomm-click-stream/e-shop-clothing.csv`,
      {}
    );

    console.log("Global...");
    await new ImportTableCommand().run(
      {
        projectPath: project,
        profileWithUpdate: true,
      },
      `${project}/example-assets/data/global-landslide-catalog/global-landslide-catalog.csv`,
      {}
    );

    console.log("Internet of Things...");
    await new ImportTableCommand().run(
      {
        projectPath: project,
        profileWithUpdate: true,
      },
      `${project}/example-assets/data/iot-env-sensor/iot-telemetry-data.csv`,
      {}
    );

    console.log("Importing example SQL transformations into the project...");
    execSync(`mv -v ${project}/example-assets/models/* ${project}/models`, {
      stdio: "inherit",
    });

    console.log("Cleaning up the project...");
    execSync(`mkdir  ${project}/data`, {
      stdio: "inherit",
    });

    execSync(`mv -v ${project}/example-assets/data ${project}`, {
      stdio: "inherit",
    });
    execSync(`rm -rf ${project}/example-assets`, {
      stdio: "inherit",
    });
    execSync(`rm -rf ${project}/example-assets.zip`, {
      stdio: "inherit",
    });
    execSync(`rm -rf ${project}/__MACOSX`, {
      stdio: "inherit",
    });
    execSync(`rm ${project}/models/query_1.sql`, {
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
