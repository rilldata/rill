import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";
import { EntityType, StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { isPortOpen } from "$common/utils/isPortOpen";
import { waitUntil } from "$common/utils/waitUtils";

interface ImportTableCommandOptions {
    project?: string;
    name?: string;
    delimiter?: string;
}

export class ImportTableCommand extends DataModelerCliCommand {
    public getCommand(): Command {
        return this.applyCommonSettings(
            new Command("import-table"),
            "Import a table from the supplied table source file."
        )
            .argument("<tableSourceFile>", "File to import the table from. Supported file types: .parquet, .csv, .tsv.")
            .option("--name <tableName>", "Optional name of the table. Defaults to file name without the folder path and extension.")
            .option("--delimiter <delimiter>", "Optional delimiter for csv and tsv files. " +
                "This is auto detected by DuckDB, but can be forced to a different delimiter if the auto-detection is not satisfactory.")
            .action((tableSourceFile, opts, command: Command) => {
                const options: ImportTableCommandOptions = command.optsWithGlobals();
                return this.run({projectPath: options.project}, tableSourceFile, options);
            });
    }

    protected async sendActions(tableSourceFile: string, {name, delimiter}: ImportTableCommandOptions): Promise<void> {
        await this.dataModelerService.dispatch("addOrUpdateTableFromFile",
            [tableSourceFile, name, {csvDelimiter: delimiter}]);
        let createdTable;
        await waitUntil(() => {
            createdTable = this.dataModelerStateService
                .getEntityStateService(EntityType.Table, StateType.Persistent)
                .getByField("path", tableSourceFile);
            return !!createdTable;
        });
        if (createdTable) {
            console.log(`Successfully imported ${tableSourceFile} into table ${createdTable.tableName}`);
        } else {
            // actual error would be printed by addOrUpdateTableFromFile
            console.log(`Failed to import table from file ${tableSourceFile}`);
        }
    }
}
