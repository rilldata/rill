import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { asyncWait, waitUntil } from "$common/utils/waitUtils";
import { getSourceNameFromFile } from "$lib/util/extract-source-name";
import { cliConfirmation } from "$common/utils/cliConfirmation";
import type { PersistentSourceEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentSourceEntityService";
import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";

interface ImportSourceCommandOptions {
  project?: string;
  name?: string;
  delimiter?: string;
  force?: boolean;
}

export class ImportSourceCommand extends DataModelerCliCommand {
  public getCommand(): Command {
    return this.applyCommonSettings(
      new Command("import-source"),
      "Import a source from the supplied file."
    )
      .argument(
        "<sourceFile>",
        "File to import the source from. Supported file types: .parquet, .csv, .tsv."
      )
      .option(
        "--name <sourceName>",
        "Optional name of the source. Defaults to file name without the folder path and extension."
      )
      .option(
        "--delimiter <delimiter>",
        "Optional delimiter for csv and tsv files. " +
          "This is auto detected by DuckDB, but can be forced to a different delimiter if the auto-detection is not satisfactory."
      )
      .option(
        "--force",
        "Option to force overwrite if the source already exists. " +
          "Without this, there will be a prompt to overwrite the source if it exists."
      )
      .action((sourceFile, opts, command: Command) => {
        const options: ImportSourceCommandOptions = command.optsWithGlobals();
        return this.run({ projectPath: options.project }, sourceFile, options);
      });
  }

  protected async sendActions(
    sourceFile: string,
    importOptions: ImportSourceCommandOptions
  ): Promise<void> {
    await this.waitIfClient();
    const sourceName = getSourceNameFromFile(sourceFile, importOptions.name);
    const existingSource = this.dataModelerStateService
      .getEntityStateService(EntityType.Source, StateType.Persistent)
      .getByField("sourceName", sourceName);

    if (existingSource && importOptions.force) {
      console.log(
        `There is already an imported source name : ${sourceName}. ` +
          "\nnForcing an overwrite." +
          `\nDropping the existing ${sourceName} from ${existingSource.path} and importing ${sourceFile}`
      );
    } else if (existingSource && !importOptions.force) {
      const shouldOverwrite = await cliConfirmation(
        `There is already an imported source name : ${sourceName}. ` +
          "\nImporting again will drop the existing source and import this one. " +
          "\nAre you sure you want to do this ? (y/N)"
      );
      if (!shouldOverwrite) return;
      console.log(
        `Dropping the existing ${sourceName} from ${existingSource.path} and importing ${sourceFile}`
      );
    }

    await this.importSource(sourceFile, importOptions, existingSource);
  }

  private async importSource(
    sourceFile: string,
    { name, delimiter }: ImportSourceCommandOptions,
    existingSource: PersistentSourceEntity
  ) {
    await this.waitIfClient();
    const sourceName = getSourceNameFromFile(sourceFile, name);
    const response = await this.dataModelerService.dispatch(
      "addOrUpdateSourceFromFile",
      [sourceFile, name, { csvDelimiter: delimiter }]
    );

    if (response.status === ActionStatus.Failure) {
      response.messages.forEach((message) => console.log(message.message));
      console.log(
        `Failed to import source ${sourceName} from file ${sourceFile}`
      );
      return;
    }

    await this.waitIfClient();

    let createdSource: PersistentSourceEntity;
    await waitUntil(() => {
      createdSource = this.dataModelerStateService
        .getEntityStateService(EntityType.Source, StateType.Persistent)
        .getByField("sourceName", sourceName);
      return !!createdSource;
    }, this.config.state.syncInterval * 5);

    if (
      (existingSource &&
        createdSource &&
        existingSource.lastUpdated < createdSource.lastUpdated) ||
      (!existingSource && createdSource)
    ) {
      console.log(
        `Successfully imported ${sourceFile} into source ${createdSource.sourceName}`
      );
    } else {
      console.log(
        `Failed to import source ${sourceName} from file ${sourceFile}`
      );
    }
  }

  private async waitIfClient() {
    if (this.isClient) await asyncWait(this.config.state.syncInterval * 2);
  }
}
