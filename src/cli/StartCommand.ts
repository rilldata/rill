import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";
import { ExpressServer } from "$server/ExpressServer";

export class StartCommand extends DataModelerCliCommand {
  public getCommand(): Command {
    return this.applyCommonSettings(
      new Command("start"),
      "Start the data-modeler UI."
    ).action((opts, command: Command) => {
      const { project } = command.optsWithGlobals();
      return this.run({
        projectPath: project,
        shouldInitState: false,
        shouldSkipDatabase: false,
        profileWithUpdate: true,
      });
    });
  }

  protected sendActions(): Promise<void> {
    return new ExpressServer(
      this.config,
      this.dataModelerService,
      this.dataModelerStateService,
      this.notificationService,
      this.metricsService
    ).init();
  }

  protected async teardown(): Promise<void> {
    // do not teardown as this will have a perpetual server
  }
}
