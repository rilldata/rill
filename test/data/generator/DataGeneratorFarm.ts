import workerpool from "workerpool";
import os from "os";
import {BATCH_SIZE, DATA_FOLDER} from "./data-constants";
import type { DataWriter } from "./writers/DataWriter";
import { ParquetFileWriter } from "./writers/ParquetFileWriter";
import { CSVFileWriter } from "./writers/CSVFileWriter";

const OUTPUT_FOLDER = `${__dirname}/../../../${DATA_FOLDER}`;
const CPU_COUNT = os.cpus().length;

export class DataGeneratorFarm {
    private readonly pool;

    public constructor(worker: string) {
        this.pool = workerpool.pool(worker);
    }

    public async generate(type: string, count: number): Promise<void> {
        console.log(`Generating ${type}`);

        const writers: Array<DataWriter> = [
            new ParquetFileWriter(type, OUTPUT_FOLDER),
            new CSVFileWriter(type, OUTPUT_FOLDER),
        ];
        for (const writer of writers) {
            await writer.init();
        }

        const handleResponse = async (rows: Array<Record<string, any>>) => {
            await Promise.all(writers.map(writer => writer.write(rows)));
        };

        for (let ids = 0; ids < count;) {
            const promises = [];
            for (let batch = 0; batch < (2 * CPU_COUNT) && ids < count; batch++, ids += BATCH_SIZE) {
                promises.push(this.pool.exec("generate", [type, ids]).then(handleResponse));
            }
            await Promise.all(promises);
        }

        for (const writer of writers) {
            await writer.close();
        }
    }

    public stop(): void {
        this.pool.terminate();
    }
}
