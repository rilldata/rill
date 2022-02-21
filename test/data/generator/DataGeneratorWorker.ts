import {DATA_GENERATOR_TYPE_MAP} from "./types/DataGeneratorTypeMap";
import {BATCH_SIZE} from "./data-constants";

export class DataGeneratorWorker {
    public generate(type: string, startId: number): Array<Record<string, any>> {
        const rows = [];
        const generatorType = DATA_GENERATOR_TYPE_MAP[type];
        for (let i = 0; i < BATCH_SIZE; i++) {
            const row = generatorType.generateRow(startId + i);
            if (row === null) break;
            rows.push(row);
        }
        return rows;
    }
}
