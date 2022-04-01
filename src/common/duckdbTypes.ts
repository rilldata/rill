export interface DuckDBColumnSummary {
    column_name: string;
    column_type: string;
    min: string;
    max: string;
    approx_unique: string;
    avg: string;
    std: string;
    q25: string;
    q50: string;
    q75: string;
    count: number;
    null_percentage: string;
}
export type DuckDBTableSummary = Array<DuckDBColumnSummary>;
