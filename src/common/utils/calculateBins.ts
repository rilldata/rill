import type {NumericHistogramBin} from "$lib/types";

export function calculateBins(data: Array<Record<string, any>>, field: string): NumericHistogramBin[] {
    // https://en.wikipedia.org/wiki/Freedman%E2%80%93Diaconis_rule
    const IQR = data[Math.round(3 * data.length / 4)][field] - data[Math.round(data.length / 4)][field];
    const binWidth = 2 * IQR / Math.cbrt(data.length);

    let curBin: NumericHistogramBin = {
        count: 1,
        low: data[0][field],
        high: data[0][field],
        bucket: 0,
    };
    const bins: NumericHistogramBin[] = [curBin];

    for (let i = 1; i < data.length - 1; i++) {
        const value = data[i][field];
        if (curBin.low + binWidth < value) {
            curBin = {
                count: 1,
                low: value,
                high: value,
                bucket: bins.length,
            };
            bins.push(curBin);
        } else {
            curBin.count++;
            curBin.high = value;
        }
    }

    return bins;
}
