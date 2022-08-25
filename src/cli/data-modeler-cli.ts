#!/usr/bin/env node
import "../moduleAlias";

import { DropTableCommand } from "$cli/DropTableCommand";
import { ExampleProjectCommand } from "$cli/ExampleProjectCommand";
import { ImportTableCommand } from "$cli/ImportTableCommand";
import { InfoCommand } from "$cli/InfoCommand";
import { InitCommand } from "$cli/InitCommand";
import { StartCommand } from "$cli/StartCommand";
import { Command } from "commander";
import { readFileSync } from "fs";

let PACKAGE_JSON = "";
try {
  PACKAGE_JSON = __dirname + "/../../package.json";
} catch (err) {
  PACKAGE_JSON = "package.json";
}
const packageJson = JSON.parse(readFileSync(PACKAGE_JSON).toString());

const program = new Command();

program
  .name("rill")
  .description("Rill Developer CLI.")
  .version(packageJson.version, "-v, --version", "Output the current version.")
  // Override help to add a capital D for display.
  .helpOption("-h, --help", "Displays help for all commands. ")
  .addHelpCommand("help [command]", "Displays help for a specific command.")
  // common across all commands
  .option(
    "--project <projectPath>",
    "Optionally indicate the path to your project. This path defaults to the current directory."
  )
  .option(
    "-d, --dev",
    "Optionally indicate if the cli is used for development purposes"
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
