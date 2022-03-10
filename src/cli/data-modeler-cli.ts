#!/usr/bin/env node --type "commonjs"
import "../moduleAlias";
import { Command } from "commander";
import { InitCommand } from "$cli/InitCommand";
import { ImportTableCommand } from "$cli/ImportTableCommand";
import { StartCommand } from "$cli/StartCommand";
import { InfoCommand } from "$cli/InfoCommand";

const program = new Command();

program
    .name("data-modeler")
    .description("Data Modeler CLI.")
    // Override help to add a capital D for display.
    .helpOption("-h, --help", "Display help for command.")
    .addHelpCommand("help [command]", "Display help for command.")
    // common across all commands
    .option("--project <projectPath>", "Optional path of project. Defaults to current directory.");

[InitCommand, ImportTableCommand, StartCommand, InfoCommand].forEach(
    CommandClass => program.addCommand(new CommandClass().getCommand())
);

program.parse();
