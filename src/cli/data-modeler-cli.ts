#!/usr/bin/env node
import "../moduleAlias";

import { DropTableCommand } from "$cli/DropTableCommand";
import { ExampleProjectCommand } from "$cli/ExampleProjectCommand";
import { ImportTableCommand } from "$cli/ImportTableCommand";
import { InfoCommand } from "$cli/InfoCommand";
import { InitCommand } from "$cli/InitCommand";
import { StartCommand } from "$cli/StartCommand";
import { Command } from "commander";
import { version } from "../../package.json";

const program = new Command();

program
  .name("rill-developer")
  .description("Rill Developer CLI.")
  .version(version, "-v, --version", "Output the current version.")
  // Override help to add a capital D for display.
  .helpOption("-h, --help", "Displays help for all commands. ")
  .addHelpCommand("help [command]", "Displays help for a specific command.")
  // common across all commands
  .option(
    "--project <projectPath>",
    "Optionally indicate the path to your project. This path defaults to the current directory."
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
