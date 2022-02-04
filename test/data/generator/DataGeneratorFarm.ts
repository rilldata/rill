import workerFarm, {Workers} from "worker-farm";
import {waitUntil} from "$common/utils/waitUtils";
import {BATCH_SIZE, DATA_GENERATOR_TYPE_MAP} from "./DataGeneratorTypeMap";
import parquet from "parquetjs";
import {execSync} from "node:child_process";
import os from "os";

const PARQUET_FOLDER = `${__dirname}/../../../`;
const CPU_COUNT = os.cpus().length;

export class DataGeneratorFarm {
    private readonly workers: Workers;

    public constructor(worker: string) {
        this.workers = workerFarm({
            maxConcurrentCallsPerWorker: 1,
        }, require.resolve(worker), ["generate"]);
    }

    public async generate(type: string, count: number): Promise<void> {
        let requests = 0;
        let responses = 0;

        execSync(`rm ${PARQUET_FOLDER}/${type}.parquet | true`);
        const parquetWriter = await parquet.ParquetWriter.openFile(
            new parquet.ParquetSchema(DATA_GENERATOR_TYPE_MAP[type].getParquetSchema()),
            `${PARQUET_FOLDER}/${type}.parquet`);

        const handleResponse = async (rows: Array<Record<string, any>>) => {
            await Promise.all(rows.map(row => parquetWriter.appendRow(row)));
            responses++;
        };

        for (let i = 0; i < count; i += BATCH_SIZE) {
            await waitUntil(() => requests - responses < 2 * CPU_COUNT);
            if (requests % 5 * CPU_COUNT === 0) console.log(`Generating ${i} rows`);
            requests++;
            this.generateInWorker(type, i).then(handleResponse);
        }

        await waitUntil(() => responses === requests);

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
