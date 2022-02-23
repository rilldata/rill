export abstract class DataWriter {
    public constructor(protected readonly type: string, protected readonly outputFolder: string) {}

    public abstract init(): Promise<void>;
    public abstract write(rows: Array<Record<string, any>>): Promise<void>;
    public abstract close(): Promise<void>;
}
