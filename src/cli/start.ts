import { Command } from "commander";

export function getStartCommand(): Command {
    return new Command("start")
        .description("Starts the data-modeler UI.")
        .action(() => {
            // TODO
        });
}
