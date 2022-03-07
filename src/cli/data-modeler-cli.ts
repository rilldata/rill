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
  .description("Data Modeler CLI.");

[InitCommand, ImportTableCommand, StartCommand, InfoCommand].forEach(
    CommandClass => program.addCommand(new CommandClass().getCommand())
);

program.parse();
