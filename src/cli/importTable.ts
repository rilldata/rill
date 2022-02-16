import { Command } from "commander";
import { getCliInstances, SAVED_STATE_FILE } from "$cli/cliFactory";
import { writeFileSync } from "fs";
import { displayProjectInfo } from "$cli/info";

export interface ImportTableOptions {
    project?: string;
    name?: string;
}

export async function importTable(tableSourceFile: string, {project, name}: ImportTableOptions): Promise<void> {
    const computedProjectPath = project ?? process.cwd();
    const {dataModelerService, dataModelerStateService} = await getCliInstances(computedProjectPath);
    await dataModelerService.dispatch("addOrUpdateTable", [tableSourceFile]);
    displayProjectInfo(computedProjectPath, dataModelerStateService.getCurrentState());
    writeFileSync(`${computedProjectPath}/${SAVED_STATE_FILE}`, JSON.stringify(dataModelerStateService.getCurrentState()));
}

export function getImportTableCommand(): Command {
    return new Command("import-table")
        .description("Import a table from the supplied table source file.")
        .argument("<tableSourceFile>", "File to import the table from. Supported types: parquet.")
        .option("--project <projectPath>", "Optional path of project. Defaults to current directory.")
        .option("--name <tableName>", "Optional name of the table. Defaults to file name without extension.")
        .action((tableSourceFile, options: ImportTableOptions) => {
            return importTable(tableSourceFile, options);
        });
}
