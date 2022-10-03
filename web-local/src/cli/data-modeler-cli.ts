#!/usr/bin/env node
import "../moduleAlias";

import { DropTableCommand } from "./DropTableCommand";
import { ExampleProjectCommand } from "./ExampleProjectCommand";
import { ImportTableCommand } from "./ImportTableCommand";
import { InfoCommand } from "./InfoCommand";
import { InitCommand } from "./InitCommand";
import { StartCommand } from "./StartCommand";
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
