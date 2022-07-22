#!/usr/bin/env node
import "../moduleAlias";
import { Command } from "commander";
import { InitCommand } from "$cli/InitCommand";
import { ImportTableCommand } from "$cli/ImportTableCommand";
import { StartCommand } from "$cli/StartCommand";
import { InfoCommand } from "$cli/InfoCommand";
import { DropTableCommand } from "$cli/DropTableCommand";
import { ExampleProjectCommand } from "$cli/ExampleProjectCommand";

const program = new Command();

program
  .name("rill")
  .description("Rill Developer CLI.")
  // Override help to add a capital D for display.
  .helpOption("-h, --help", "Displays help for all commands. ")
  .addHelpCommand("help [command]", "Displays help for a specific command. ")
  // common across all commands
  .option(
    "--project <projectPath> ",
    "Optionally indicate the path to your project. This path defaults to the current directory. "
  );

[
  InitCommand,
  ImportTableCommand,
  StartCommand,
  DropTableCommand,
  InfoCommand,
  ExampleProjectCommand,
].forEach((CommandClass) =>
  program.addCommand(new CommandClass().getCommand())
);

program.parse();

process.on("uncaughtException", (error) => console.error(error));
process.on("unhandledRejection", (error) => console.error(error));
