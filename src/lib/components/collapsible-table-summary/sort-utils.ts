export const categoricals = new Set(['BYTE_ARRAY', 'VARCHAR']);

export function sortByCardinality(a,b) {
    if (a.summary && b.summary) {
        if (a.summary.cardinality < b.summary.cardinality) {
            return 1;
        } else if (a.summary.cardinality > b.summary.cardinality) {
            return -1;
        } else {
            return sortByName(a,b);
        }
    } else {
        return 0;
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
    if (categoricals.has(a.type) && !categoricals.has(b.type)) return 1;
    if (!categoricals.has(a.type) && categoricals.has(b.type)) return -1;
    if ((a.conceptualType === 'TIMESTAMP' || a.type === 'TIMESTAMP') && (b.conceptualType !== 'TIMESTAMP' && b.type !== 'TIMESTAMP')) {
                return -1;
    } else if ((a.conceptualType !== 'TIMESTAMP' && a.type !== 'TIMESTAMP') && (b.conceptualType === 'TIMESTAMP' || b.type ==='TIMESTAMP')) {
        return 1;
    }
    return 0;
}

export function sortByName(a,b) {
    return (a.name > b.name) ? 1 : -1;
}

export function defaultSort(a, b) {
    const byType = sortByType(a,b);
    if (byType !== 0) return byType;
    if (categoricals.has(a.type) && !categoricals.has(b.type)) return sortByNullity(b,a);
    return sortByCardinality(a,b);
}