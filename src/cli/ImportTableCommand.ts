import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";
import { InfoCommand } from "$cli/InfoCommand";

interface ImportTableOptions {
    project?: string;
    name?: string;
}

export class ImportTableCommand extends DataModelerCliCommand {
    public getCommand(): Command {
        return new Command("import-table")
            .description("Import a table from the supplied table source file.")
            .argument("<tableSourceFile>", "File to import the table from. Supported types: parquet.")
            .option("--project <projectPath>", "Optional path of project. Defaults to current directory.")
            .option("--name <tableName>", "Optional name of the table. Defaults to file name without extension.")
            .action((tableSourceFile, options: ImportTableOptions) => {
                return this.run(options.project, tableSourceFile, options);
            });
    }

    protected async sendActions(tableSourceFile: string, {name}: ImportTableOptions): Promise<void> {
        await this.dataModelerService.dispatch("addOrUpdateTable", [tableSourceFile]);
        InfoCommand.displayProjectInfo(this.projectPath, this.dataModelerStateService.getCurrentState());
    }
}
