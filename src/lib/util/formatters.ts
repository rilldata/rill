import { INTERVALS, INTEGERS, FLOATS, CATEGORICALS, TIMESTAMPS, BOOLEANS, PreviewRollupInterval } from "$lib/duckdb-data-types";
import type { Interval } from "$lib/duckdb-data-types";
import { format } from "d3-format";
import { timeFormat } from "d3-time-format";

const zeroPad = format('02d');
const msPad = format('03d')
export const formatInteger = format(',');
const formatRate = format('.1f');

/**
 * changes precision depending on the 
 */
export function formatBigNumberPercentage(v) {
    if (v < .0001) {
        const f = format('.4%')(v);
        if (f === '0.0000%') {
            return "~ 0%"
        } else {
            return f
        }
    } else {
        return format('.2%')(v);
    }
}

export function removeTimezoneOffset(dt) {
    return new Date(dt.getTime() + dt.getTimezoneOffset() * 60000)
}

export const standardTimestampFormat = (v, type = 'TIMESTAMP') => {
    let fmt = timeFormat('%b %d, %Y %I:%M:%S');
    if (type === 'DATE') {
        fmt = timeFormat('%b %d, %Y');
    }
    return fmt(removeTimezoneOffset(new Date(v)));
}

export const fullTimestampFormat = (v) => {
    let fmt = timeFormat('%b %d, %Y %I:%M:%S.%L');
    return fmt(removeTimezoneOffset(new Date(v)));
}

export const datePortion = timeFormat('%b %d, %Y');
export const timePortion = timeFormat("%I:%M:%S");

export function microsToTimestring(microseconds:number) {
    // to format micros, we need to translate this to hh:mm:ss.
    // start with hours/
    let sign = Math.sign(microseconds);
    let micros = Math.abs(microseconds);
    const hours = ~~(micros / 1000 / 1000 / 60 / 60);
    let remaining = micros - (hours * 1000 * 1000 * 60 * 60);
    const minutes = ~~(remaining / 1000/ 1000 / 60);
    //const seconds = (remaining - (minutes * 1000 * 1000 * 60)) / 1000 / 1000;
    remaining -= (minutes * 1000 * 1000 * 60);
    const seconds = ~~(remaining / 1000 / 1000);
    remaining -= (seconds * 1000 * 1000);
    const ms = ~~(remaining / 1000);
    if (hours === 0 && minutes === 0 && seconds === 0 && ms > 0) {
        return `${sign == 1 ? '' : '-'}${ms}ms`
    }
    return `${sign == 1 ? '' : '-'}${zeroPad(hours)}:${zeroPad(minutes)}:${zeroPad(seconds)}.${msPad(ms)}`
}

export function intervalToTimestring(interval:Interval) {
    const months = interval.months ? `${formatInteger(interval.months)} month${interval.months > 1 ? 's' : ''} ` : '';
    const days = interval.days ? `${formatInteger(interval.days)} day${interval.days > 1 ? 's' : ''} ` : '';
    const time = (interval.months > 0 || interval.days > 1) ? '' : microsToTimestring(interval.micros);
    // if only days && days > 365, convert to years?
    if (interval.months === 0 && interval.days > 0 && interval.days > 365) return `${formatRate(interval.days / 365)} years`
    return `${months}${days}${time}`;
}

export function formatCompactInteger(n:number) {
    let fmt:Function;
    if (n <= 1000) {
        fmt = formatInteger;
        return fmt(~~n);
    } else {
        fmt = format('.3s');
        return fmt(n);
    }
}

export function formatDataType(value:any, type:string) {
    if (INTEGERS.has(type)) {
        return value;
    } else if (FLOATS.has(type)) {
        return value;
    } else if (CATEGORICALS.has(type)) {
        return value;
    } else if (TIMESTAMPS.has(type)) {
        return standardTimestampFormat(value, type);
    } else if (INTERVALS.has(type)) {
        return intervalToTimestring(value);
    } else if (BOOLEANS.has(type)) {
        return value;
    }
}

/** These will be used in the string */
export const PreviewRollupIntervalFormatter = {
    [PreviewRollupInterval.ms]: 'millisecond-level',         /** showing rows binned by ms */
    [PreviewRollupInterval.second]: 'second-level', /** showing rows binned by second */
    [PreviewRollupInterval.minute]: 'minute-level', /** showing rows binned by minute */
    [PreviewRollupInterval.hour]: 'hourly',          /** showing hourly counts */
    [PreviewRollupInterval.day]: 'daily',            /** showing daily counts */
    [PreviewRollupInterval.month]: 'monthly',        /** showing monthly counts */
    [PreviewRollupInterval.year]: 'yearly',          /** showing yearly counts */
}