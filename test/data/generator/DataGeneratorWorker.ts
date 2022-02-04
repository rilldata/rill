import {BATCH_SIZE, DATA_GENERATOR_TYPE_MAP} from "./DataGeneratorTypeMap";

export class DataGeneratorWorker {
    public generate(type: string, startId: number): Array<Record<string, any>> {
        const rows = [];
        const generatorType = DATA_GENERATOR_TYPE_MAP[type];
        for (let i = 0; i < BATCH_SIZE; i++) {
            rows.push(generatorType.generateRow(startId + i));
        }
        return rows;
    }
}
