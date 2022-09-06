import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { DimensionsActions } from "$common/rill-developer-service/DimensionsActions";
import { MeasuresActions } from "$common/rill-developer-service/MeasuresActions";
import { MetricsDefinitionActions } from "$common/rill-developer-service/MetricsDefinitionActions";
import { MetricsViewActions } from "$common/rill-developer-service/MetricsViewActions";
import { RillDeveloperService } from "$common/rill-developer-service/RillDeveloperService";
import { ExpressServer } from "$server/ExpressServer";
import { Command } from "commander";

export class StartCommand extends DataModelerCliCommand {
  public getCommand(): Command {
    return this.applyCommonSettings(
      new Command("start"),
      "Start the Rill Developer application. "
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

  protected async sendActions(): Promise<void> {
    const activeEntity =
      this.dataModelerStateService.getApplicationState().activeEntity;
    if (!activeEntity || activeEntity?.type !== EntityType.Model) {
      // set dummy asset as active to show onboarding steps
      await this.dataModelerService.dispatch("setActiveAsset", [
        EntityType.Model,
        undefined,
      ]);
    }

    return new ExpressServer(
      this.config,
      this.dataModelerService,
      new RillDeveloperService(
        this.dataModelerStateService,
        this.dataModelerService,
        this.dataModelerService.getDatabaseService(),
        [
          MetricsDefinitionActions,
          DimensionsActions,
          MeasuresActions,
          MetricsViewActions,
        ].map(
          (RillDeveloperActionsClass) =>
            new RillDeveloperActionsClass(
              this.config,
              this.dataModelerStateService,
              this.dataModelerService.getDatabaseService()
            )
        )
      ),
      this.dataModelerStateService,
      this.notificationService,
      this.metricsService
    ).init();
  }

  protected async teardown(): Promise<void> {
    // do not teardown as this will have a perpetual server
  }
}
