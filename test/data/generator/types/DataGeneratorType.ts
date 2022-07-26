export type ParquetSchemaType = Record<string, ParquetDataType>;
export interface ParquetDataType {
  type: string;
  optional?: boolean;
  repeated?: boolean;
  fields?: ParquetSchemaType;
}

export abstract class DataGeneratorType {
  public abstract generateRow(id: number): Record<string, unknown>;

  public abstract getParquetSchema(): ParquetSchemaType;

  public csvExtension = "csv";
  public csvDelimiter = ",";
  public columnsOrder: Array<string>;

  protected generateRandomInt(min: number, max: number): number {
    return Math.round(this.generateRandomFloat(min, max));
  }

  protected generateRandomFloat(min: number, max: number): number {
    return Number((Math.random() * (max - min) + min).toFixed(2));
  }

  protected generateRandomTimestamp(
    startDate: string,
    endDate: string
  ): string {
    const startDateNum = new Date(`${startDate} UTC`).getTime();
    const endDateNum = new Date(`${endDate} UTC`).getTime();

    return new Date(
      this.generateRandomInt(startDateNum, endDateNum)
    ).toISOString();
  }

  protected selectRandomEntry<T>(entries: Array<T>): T {
    const index = Math.floor(Math.random() * entries.length);
    return entries[index];
  }
}
