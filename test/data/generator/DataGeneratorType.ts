export type ParquetSchemaType = Record<string, ParquetDataType>;
export interface ParquetDataType {
    type: string;
    optional?: boolean;
    repeated?: boolean;
    fields?: ParquetSchemaType;
}

export abstract class DataGeneratorType {
    public abstract generateRow(id: number): Record<string, any>;

    public abstract getParquetSchema(): ParquetSchemaType;

    protected generateRandomInt(min: number, max: number): number {
        return Math.round(this.generateRandomFloat(min, max));
    }

    protected generateRandomFloat(min: number, max: number): number {
        return Number((Math.random() * (max - min) + min).toFixed(2));
    }

    protected generateRandomTimestamp(startDate: string, endDate: string): number {
        const startDateNum = new Date(`${startDate} UTC`).getTime();
        const endDateNum = new Date(`${endDate} UTC`).getTime();

        return this.generateRandomInt(startDateNum, endDateNum);
    }

    protected selectRandomEntry<T>(entries: Array<T>): T {
        const index = Math.floor(Math.random() * entries.length);
        return entries[index];
    }
}
