#!/usr/bin/env node
import "../moduleAlias";
import { Command } from "commander";
import { InitCommand } from "$cli/InitCommand";
import { ImportSourceCommand } from "$cli/ImportSourceCommand";
import { StartCommand } from "$cli/StartCommand";
import { InfoCommand } from "$cli/InfoCommand";
import { DropSourceCommand } from "$cli/DropSourceCommand";
import { ExampleProjectCommand } from "$cli/ExampleProjectCommand";

const program = new Command();

program
  .name("rill-developer")
  .description("Rill Developer CLI.")
  // Override help to add a capital D for display.
  .helpOption("-h, --help", "Display help for command.")
  .addHelpCommand("help [command]", "Display help for command.")
  // common across all commands
  .option(
    "--project <projectPath>",
    "Optional path of project. Defaults to current directory."
  );

[
  InitCommand,
  ImportSourceCommand,
  StartCommand,
  DropSourceCommand,
  InfoCommand,
  ExampleProjectCommand,
].forEach((CommandClass) =>
  program.addCommand(new CommandClass().getCommand())
);

program.parse();

process.on("uncaughtException", (error) => console.error(error));
process.on("unhandledRejection", (error) => console.error(error));
