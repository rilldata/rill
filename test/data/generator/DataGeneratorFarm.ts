import workerFarm, {Workers} from "worker-farm";
import {DATA_GENERATOR_TYPE_MAP} from "./DataGeneratorTypeMap";
import parquet from "parquetjs";
import {execSync} from "node:child_process";
import os from "os";
import {BATCH_SIZE} from "./data-constants";

const PARQUET_FOLDER = `${__dirname}/../../..`;
const CPU_COUNT = os.cpus().length;

export class DataGeneratorFarm {
    private readonly workers: Workers;

    public constructor(worker: string) {
        this.workers = workerFarm({
            maxConcurrentCallsPerWorker: 1,
        }, require.resolve(worker), ["generate"]);
    }

    public async generate(type: string, count: number): Promise<void> {
        console.log(`Generating ${type}`);

        const parquetFile = `${PARQUET_FOLDER}/${type}.parquet`;
        execSync(`rm ${parquetFile} | true`);
        const parquetWriter = await parquet.ParquetWriter.openFile(
            new parquet.ParquetSchema(DATA_GENERATOR_TYPE_MAP[type].getParquetSchema()), parquetFile);
        parquetWriter.setRowGroupSize(2 * CPU_COUNT * BATCH_SIZE);

        const handleResponse = async (rows: Array<Record<string, any>>) => {
            await Promise.all(rows.map(row => parquetWriter.appendRow(row)));
        };

        for (let ids = 0; ids < count;) {
            const promises = [];
            for (let batch = 0; batch < (2 * CPU_COUNT) && ids < count; batch++, ids += BATCH_SIZE) {
                promises.push(this.generateInWorker(type, ids).then(handleResponse));
            }
            await Promise.all(promises);
        }

        await parquetWriter.close();
    }

    public stop(): void {
        workerFarm.end(this.workers);
    }

    private generateInWorker(type: string, startId: number) {
        return new Promise((resolve, reject) => {
            this.workers.generate([type, startId], (err, resp) => {
                if (err) reject(err);
                else resolve(resp);
            });
        });
    }
}
