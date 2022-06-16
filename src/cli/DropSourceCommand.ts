import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";
import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";

export class DropSourceCommand extends DataModelerCliCommand {
  public getCommand(): Command {
    return this.applyCommonSettings(
      new Command("drop-source"),
      "Drop a source."
    )
      .argument("<sourceName>", "Name of the source to drop.")
      .action((sourceName, opts, command: Command) => {
        const { project } = command.optsWithGlobals();
        return this.run({ projectPath: project }, sourceName);
      });
  }

  protected async sendActions(sourceName: string): Promise<void> {
    const response = await this.dataModelerService.dispatch("dropSource", [
      sourceName,
    ]);
    if (response.status === ActionStatus.Failure) {
      response.messages.forEach((message) => console.log(message.message));
      console.log(`Failed to drop source ${sourceName}`);
      return;
    }
    console.log(`Successfully dropped source ${sourceName}`);
  }
}
