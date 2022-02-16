#!/usr/bin/env node

import "../moduleAlias";
import { Command } from "commander";
import { getInitCommand } from "./init";
import { getImportTableCommand } from "./importTable";
import { getDropTableCommand } from "./dropTable";
import { getStartCommand } from "./start";
import { getInfoCommand } from "$cli/info";
const program = new Command();

program
  .name("data-modeler")
  .description("Data Modeler CLI.");

program.addCommand(getInitCommand());
program.addCommand(getImportTableCommand());
program.addCommand(getDropTableCommand());
program.addCommand(getStartCommand());
program.addCommand(getInfoCommand());

program.parse();
