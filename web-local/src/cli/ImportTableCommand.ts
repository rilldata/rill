import { DataModelerCliCommand } from "./DataModelerCliCommand";
import { Command } from "commander";
import {
  EntityType,
  StateType,
} from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  asyncWait,
  waitUntil,
} from "@rilldata/web-local/common/utils/waitUtils";
import { getTableNameFromFile } from "@rilldata/web-local/lib/util/extract-table-name";
import { cliConfirmation } from "@rilldata/web-local/common/utils/cliConfirmation";
import type { PersistentTableEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { ActionStatus } from "@rilldata/web-local/common/data-modeler-service/response/ActionResponse";

interface ImportTableCommandOptions {
  project?: string;
  name?: string;
  delimiter?: string;
  force?: boolean;
}

export class ImportTableCommand extends DataModelerCliCommand {
  public getCommand(): Command {
    return this.applyCommonSettings(
      new Command("import-source"),
      "Imports a source file into Rill Developer."
    )
      .argument(
        "<sourceFile>",
        "Specify the path to the source file to be imported. " +
          "Supported file types include .parquet, .csv, .tsv. "
      )
      .option(
        "--name <sourceName>",
        "Optional rename the source created in Rill Developer. " +
          "If no name is indicated, the source name defaults to a sanitized version of the file name. "
      )
      .option(
        "--delimiter <delimiter>",
        "Optional delimiter for csv and tsv files. " +
          "If no delimiter is indicaated, file parsing is automatically detected by DuckDB. "
      )
      .option(
        "--force",
        "Optionally force overwrite if the source name already exists in Rill Developer. " +
          "Without this option enabled, there will be a prompt to overwrite the source if it exists."
      )
      .action((tableSourceFile, opts, command: Command) => {
        const options: ImportTableCommandOptions = command.optsWithGlobals();
        return this.run(
          { projectPath: options.project },
          tableSourceFile,
          options
        );
      });
  }

  protected async sendActions(
    tableSourceFile: string,
    importOptions: ImportTableCommandOptions
  ): Promise<void> {
    await this.waitIfClient();
    const tableName = getTableNameFromFile(tableSourceFile, importOptions.name);
    const existingTable = this.dataModelerStateService
      .getEntityStateService(EntityType.Table, StateType.Persistent)
      .getByField("tableName", tableName);

    if (existingTable && importOptions.force) {
      console.log(
        `There is already a source named ${tableName}. ` +
          "\nForcing an overwrite." +
          `\nDropping the existing source ${tableName} from ${existingTable.path} and importing ${tableSourceFile}`
      );
    } else if (existingTable && !importOptions.force) {
      const shouldOverwrite = await cliConfirmation(
        `There is already a source named ${tableName}. ` +
          "\nDo you want to drop the existing source and import this one? (y/N)"
      );
      if (!shouldOverwrite) return;
      console.log(
        `Dropping the existing source ${tableName} from ${existingTable.path} and importing ${tableSourceFile}`
      );
    }

    await this.importTable(tableSourceFile, importOptions, existingTable);
  }

  private async importTable(
    tableSourceFile: string,
    { name }: ImportTableCommandOptions,
    existingTable: PersistentTableEntity
  ) {
    await this.waitIfClient();
    const tableName = getTableNameFromFile(tableSourceFile, name);
    const response = await this.dataModelerService.dispatch(
      "importTableFromCLI",
      [tableSourceFile, tableName]
    );

    if (response.status === ActionStatus.Failure) {
      response.messages.forEach((message) => console.log(message.message));
      console.log(
        `Failed to import source ${tableName} from file ${tableSourceFile}`
      );
      return;
    }

    await this.waitIfClient();

    let createdTable: PersistentTableEntity;
    await waitUntil(() => {
      createdTable = this.dataModelerStateService
        .getEntityStateService(EntityType.Table, StateType.Persistent)
        .getByField("tableName", tableName);
      return !!createdTable;
    }, this.config.state.syncInterval * 5);

    if (
      (existingTable &&
        createdTable &&
        existingTable.lastUpdated <= createdTable.lastUpdated) ||
      (!existingTable && createdTable)
    ) {
      console.log(
        `Successfully imported ${tableSourceFile} into source ${createdTable.tableName}`
      );
    } else {
      console.log(
        `Failed to import source ${tableName} from file ${tableSourceFile}`
      );
    }
  }

  private async waitIfClient() {
    if (this.isClient) await asyncWait(this.config.state.syncInterval * 2);
  }
}
