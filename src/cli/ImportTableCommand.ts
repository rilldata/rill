import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";
import { InfoCommand } from "$cli/InfoCommand";

interface ImportTableCommandOptions {
    project?: string;
    name?: string;
    delimiter?: string;
}

export class ImportTableCommand extends DataModelerCliCommand {
    public getCommand(): Command {
        return new Command("import-table")
            .description("Import a table from the supplied table source file.")
            .argument("<tableSourceFile>", "File to import the table from. Supported file types: .parquet, .csv, .tsv.")
            .option("--project <projectPath>", "Optional path of project. Defaults to current directory.")
            .option("--name <tableName>", "Optional name of the table. Defaults to file name without the folder path and extension.")
            .option("--delimiter <delimiter>", "Optional delimiter for csv and tsv files. " +
                "This is auto detected by DuckDB, but can be forced to a different delimiter if the auto-detection is not satisfactory.")
            .action((tableSourceFile, options: ImportTableCommandOptions) => {
                return this.run({projectPath: options.project}, tableSourceFile, options);
            });
    }

    protected async sendActions(tableSourceFile: string, {name, delimiter}: ImportTableCommandOptions): Promise<void> {
        await this.dataModelerService.dispatch("addOrUpdateTableFromFile",
            [tableSourceFile, name, {csvDelimiter: delimiter}]);
        console.log(`Successfully imported ${tableSourceFile}`);
        InfoCommand.displayProjectInfo(this.projectPath, this.dataModelerStateService);
    }
}
