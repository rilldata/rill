import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import { RootConfig } from "$common/config/RootConfig";
import { DatabaseConfig } from "$common/config/DatabaseConfig";
import type { Command } from "commander";
import type { SocketNotificationService } from "$common/socket/SocketNotificationService";
import { ServerConfig } from "$common/config/ServerConfig";
import { clientFactory } from "$common/clientFactory";
import { isPortOpen } from "$common/utils/isPortOpen";
import type { MetricsService } from "$common/metrics-service/MetricsService";
import { RillDeveloper } from "$server/RillDeveloper";
import { ProjectConfig } from "$common/config/ProjectConfig";
import { LocalConfig } from "$common/config/LocalConfig";

const DATABASE_NAME = "stage.db";

export interface CliRunArgs {
  projectPath: string;
  duckDbPath?: string;
  shouldInitState?: boolean;
  shouldSkipDatabase?: boolean;
  profileWithUpdate?: boolean;
}

export abstract class DataModelerCliCommand {
  protected dataModelerService: DataModelerService;
  protected dataModelerStateService: DataModelerStateService;
  protected notificationService: SocketNotificationService;
  protected metricsService: MetricsService;
  protected projectPath: string;
  protected config: RootConfig;
  protected isClient: boolean;

  protected rillDeveloper: RillDeveloper;

  private async init(cliRunArgs: CliRunArgs): Promise<void> {
    this.projectPath = cliRunArgs.projectPath ?? process.cwd();
    cliRunArgs.shouldInitState ??= true;
    cliRunArgs.shouldSkipDatabase ??= true;
    cliRunArgs.profileWithUpdate ??= false;

    this.config = new RootConfig({
      database: new DatabaseConfig({
        databaseName: DATABASE_NAME,
        skipDatabase: cliRunArgs.shouldSkipDatabase,
      }),
      server: new ServerConfig({
        serverPort: Number(process.env.RILL_SERVER_PORT ?? 8080),
        serveStaticFile: true,
      }),
      local: new LocalConfig({
        isDev: Boolean(process.env.RILL_IS_DEV ?? false),
      }),
      project: new ProjectConfig({ duckDbPath: cliRunArgs.duckDbPath }),
      projectFolder: this.projectPath,
      profileWithUpdate: cliRunArgs.profileWithUpdate,
    });

    const isServerRunning = await isPortOpen(this.config.server.socketPort);

    if (isServerRunning) {
      await this.initClientInstances();
    } else {
      // database should be started when server is not running.
      // We can write to database in this case
      this.config.database.skipDatabase = false;
      await this.initServerInstances();
    }

    this.isClient = isServerRunning;
  }

  protected async teardown(): Promise<void> {
    if (this.isClient) {
      await this.dataModelerService.destroy();
    } else {
      await this.rillDeveloper.destroy();
    }
  }

  private async initServerInstances() {
    this.rillDeveloper = RillDeveloper.getRillDeveloper(this.config);

    this.dataModelerService = this.rillDeveloper.dataModelerService;
    this.dataModelerStateService = this.rillDeveloper.dataModelerStateService;
    this.notificationService = this.rillDeveloper
      .notificationService as SocketNotificationService;
    this.metricsService = this.rillDeveloper.metricsService;

    await this.rillDeveloper.init();
  }

  private async initClientInstances() {
    const { dataModelerService, dataModelerStateService, metricsService } =
      clientFactory(this.config);
    this.dataModelerService = dataModelerService;
    this.dataModelerStateService = dataModelerStateService;
    this.metricsService = metricsService;
    await dataModelerService.init();
  }

  public async run(
    cliRunArgs: CliRunArgs,
    ...args: Array<unknown>
  ): Promise<void> {
    await this.init(cliRunArgs);
    await this.sendActions(...args);
    await this.teardown();
  }

  protected abstract sendActions(...args: Array<unknown>): Promise<void>;

  protected applyCommonSettings(
    command: Command,
    description: string
  ): Command {
    return (
      command
        .description(description)
        // override default help text to add capital D for display
        .helpOption("-h, --help", "Displays help for each command.")
        // common across all commands
        .option(
          "--project <projectPath>",
          "Optionally indicate the path to your project. This path defaults to the current directory."
        )
    );
  }

  public abstract getCommand(): Command;
}
