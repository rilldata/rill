import { BOOLEANS, CATEGORICALS, NUMERICS, TIMESTAMPS } from "$lib/duckdb-data-types";

export function sortByCardinality(a,b) {
    if (a.summary && b.summary) {
        if (a.summary.cardinality < b.summary.cardinality) {
            return 1;
        } else if (a.summary.cardinality > b.summary.cardinality) {
            return -1;
        } else {
            return sortByType(a,b);
        }
    } else {
        return sortByType(a, b);
    }
}

export function sortByNullity(a,b) {
    if (a.nullCount !== undefined && b.nullCount !== undefined) {
        if (a.nullCount < b.nullCount) {
            return 1;
        } else if ((a.nullCount > b.nullCount)) {
            return -1;
        } else {
            const byType = sortByType(a,b);
            if (byType) return byType;
            return sortByName(a,b);
        }
    }

    return sortByName(a,b);
}

export function sortByType(a,b) {
    if (BOOLEANS.has(a.type) && !BOOLEANS.has(b.type)) return 1;
    else if (!BOOLEANS.has(a.type) && BOOLEANS.has(b.type)) return -1;
    else if (CATEGORICALS.has(a.type) && !CATEGORICALS.has(b.type)) return 1;
    else if (!CATEGORICALS.has(a.type) && CATEGORICALS.has(b.type)) return -1;
    else if (NUMERICS.has(a.type) && !NUMERICS.has(b.type)) return 1;
    else if (!NUMERICS.has(a.type) && NUMERICS.has(b.type)) return -1;
    else if (TIMESTAMPS.has(a.type) && TIMESTAMPS.has(b.type)) {
                return -1;
    } else if (!TIMESTAMPS.has(a.type) && TIMESTAMPS.has(b.type)) {
        return 1;
    }
    return 0;//sortByName(a, b);
}

export function sortByName(a,b) {
    return (a.name > b.name) ? 1 : -1;
}

export function defaultSort(a, b) {
    const byType = sortByType(a,b);
    if (byType !== 0) return byType;
    if (
        CATEGORICALS.has(a.type) && !CATEGORICALS.has(b.type)
    ) return sortByNullity(b,a);
    return sortByCardinality(a,b);
}