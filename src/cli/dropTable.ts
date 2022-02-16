import { Command } from "commander";

export function getDropTableCommand(): Command {
    return new Command("drop-table")
        .description("Drops a table.")
        .argument("<tableName>", "Name of the table to drop.")
        .action((tableName) => {
            // TODO
            console.log(tableName);
        });
}
