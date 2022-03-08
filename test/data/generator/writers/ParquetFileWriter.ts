import parquet from "parquetjs";
import { execSync } from "node:child_process";
import { DATA_GENERATOR_TYPE_MAP } from "../types/DataGeneratorTypeMap";
import { BATCH_SIZE } from "../data-constants";
import os from "os";
import { DataWriter } from "./DataWriter";
import { existsSync } from "fs";

const CPU_COUNT = os.cpus().length;

export class ParquetFileWriter extends DataWriter {
    private parquetWriter;

    public async init(): Promise<void> {
        const parquetFile = `${this.outputFolder}/${this.type}.parquet`;
        if (existsSync(parquetFile)) {
            execSync(`rm ${parquetFile} | true`);
        }
        this.parquetWriter = await parquet.ParquetWriter.openFile(
            new parquet.ParquetSchema(DATA_GENERATOR_TYPE_MAP[this.type].getParquetSchema()), parquetFile);
        this.parquetWriter.setRowGroupSize(2 * CPU_COUNT * BATCH_SIZE);
    }

    public async write(rows: Array<Record<string, any>>): Promise<void> {
        for (const row of rows) {
            await this.parquetWriter.appendRow(row);
        }
    }

    public async close(): Promise<void> {
        await this.parquetWriter.close();
    }
}
