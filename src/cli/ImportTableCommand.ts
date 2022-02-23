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
            .argument("<tableSourceFile>", "File to import the table from. Supported types: parquet.")
            .option("--project <projectPath>", "Optional path of project. Defaults to current directory.")
            .option("--name <tableName>", "Optional name of the table. Defaults to file name without extension.")
            .option("--delimiter <delimiter>", "Optional name of the table. Defaults to file name without extension.")
            .action((tableSourceFile, options: ImportTableCommandOptions) => {
                return this.run(options.project, tableSourceFile, options);
            });
    }

    protected async sendActions(tableSourceFile: string, {name, delimiter}: ImportTableCommandOptions): Promise<void> {
        await this.dataModelerService.dispatch("addOrUpdateTableFromFile",
            [tableSourceFile, name, {csvDelimiter: delimiter}]);
        InfoCommand.displayProjectInfo(this.projectPath, this.dataModelerStateService.getCurrentState());
    }
}
