import { DataWriter } from "./DataWriter";
import { execSync } from "node:child_process";
import { DATA_GENERATOR_TYPE_MAP } from "../types/DataGeneratorTypeMap";
import { FileHandle, open } from "fs/promises";
import { existsSync } from "fs";

export class CSVFileWriter extends DataWriter {
  private csvWriter: FileHandle;
  private totalSize = 0;

  public async init(): Promise<void> {
    const generator = DATA_GENERATOR_TYPE_MAP[this.type];
    const csvFile = `${this.outputFolder}/${this.type}.${generator.csvExtension}`;
    if (existsSync(csvFile)) {
      execSync(`rm ${csvFile} | true`);
    }
    this.csvWriter = await open(csvFile, "w");
    await this.writeToFile(
      generator.columnsOrder.join(generator.csvDelimiter) + "\n"
    );
  }

  public async write(rows: Array<Record<string, unknown>>): Promise<void> {
    const generator = DATA_GENERATOR_TYPE_MAP[this.type];
    for (const row of rows) {
      await this.writeToFile(
        generator.columnsOrder
          .map((column) => row[column] ?? "")
          .join(generator.csvDelimiter) + "\n"
      );
    }
  }

  public async close(): Promise<void> {
    await this.csvWriter.write("", this.totalSize - 1);
    await this.csvWriter.close();
  }

  private async writeToFile(data: string): Promise<void> {
    this.totalSize += data.length;
    await this.csvWriter.write(data);
  }
}
