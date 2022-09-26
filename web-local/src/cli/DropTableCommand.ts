import { DataModelerCliCommand } from "./DataModelerCliCommand";
import { Command } from "commander";
import { ActionStatus } from "@rilldata/web-local/common/data-modeler-service/response/ActionResponse";

export class DropTableCommand extends DataModelerCliCommand {
  public getCommand(): Command {
    return this.applyCommonSettings(
      new Command("drop-source"),
      "Drops the source from Rill Developer. "
    )
      .argument("<sourceName>", "Name of the source to drop. ")
      .action((tableName, opts, command: Command) => {
        const { project } = command.optsWithGlobals();
        return this.run({ projectPath: project }, tableName);
      });
  }

  protected async sendActions(tableName: string): Promise<void> {
    const response = await this.dataModelerService.dispatch("dropTable", [
      tableName,
    ]);
    if (response.status === ActionStatus.Failure) {
      response.messages.forEach((message) => console.log(message.message));
      console.log(`Failed to drop source ${tableName}. `);
      return;
    }
    console.log(`Successfully dropped source ${tableName}. `);
  }
}
