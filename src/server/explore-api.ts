/** fundamental rollup function */

export interface Metric {
    name:string;
    field:string;
    description:string;
    function:string;
    aka?:string;
}

export interface Dimension {
    name:string;
    field:string;
    definition:string;
}

export interface DimensionSlice {
    field:string;
    value:(string | number);
}

interface RollupQueryArguments {
    table:string;
    timeGrain: 'day' | 'hour';
    timeField:string;
    metrics:Metric[];
    dimensions?:DimensionSlice[];
}

export function rollupQuery({ table, timeField, metrics, timeGrain = 'day', dimensions = [] } : RollupQueryArguments) : string {
    const timeFieldStatement = `date_trunc('${timeGrain}', ${timeField})`;
    let groupStatement = `GROUP BY ${timeFieldStatement}`;
    let whereStatement = dimensions.length 
        ? `WHERE ${dimensions.map((slice) => `${slice.field} = ${typeof slice.value === 'string' ? `'${slice.value}'` : slice.value}`).join(' AND ')}` 
        : '';
    return `
        SELECT 
            ${metrics.map(metric => `${metric.function}(${metric.field}) AS ${metric.aka || metric.field}`).join(', ')}, ${timeFieldStatement} as _ts
        FROM ${table}
        ${whereStatement}
        ${groupStatement}
        ORDER BY ${timeFieldStatement} asc
    `
}

interface TopKQueryArguments {
    table:string;
    field:string;
    metrics:Metric[];
    dimensions?:DimensionSlice[]
}

export function topKQuery({ table, field, metrics = [], dimensions = [] } : TopKQueryArguments) : string {
    let groupStatement = `GROUP BY ${field}`;
    let whereStatement = dimensions.length 
        ? `WHERE ${dimensions.map((slice) => `${slice.field} = ${typeof slice.value === 'string' ? `'${slice.value}'` : slice.value}`).join(' AND ')}` 
        : '';
    return `
    SELECT
    ${metrics.map(metric => `${metric.function}(${metric.field}) AS ${metric.aka || metric.field}`).join(', ')},
    ${field}
        FROM ${table}
    ${whereStatement}
    ${groupStatement}
    `
}