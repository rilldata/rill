import { format } from "d3-format";
import { timeFormat } from "d3-time-format";


const zeroPad = format('02d');
const formatInteger = format(',');
const formatRate = format('.1f');
export const standardTimestampFormat = timeFormat('%b %d, %Y %I:%M:%S');

export function microsToTimestring(micros:number) {
    // to format micros, we need to translate this to hh:mm:ss.
    // start with hours/
    const hours = ~~(micros / 1000 / 1000 / 60 / 60);
    let remaining = micros - (hours * 1000 * 1000 * 60 * 60);
    const minutes = ~~(remaining / 1000/ 1000 / 60);
    const seconds = (remaining - (minutes * 1000 * 1000 * 60)) / 1000 / 1000;
    return `${zeroPad(hours)}:${zeroPad(minutes)}:${zeroPad(seconds)}`
}

interface Interval {
    months:number;
    days:number;
    micros:number;
}

export function intervalToTimestring(interval:Interval) {
    const months = interval.months ? `${formatInteger(interval.months)} month${interval.months > 1 ? 's' : ''} ` : '';
    const days = interval.days ? `${formatInteger(interval.days)} day${interval.days > 1 ? 's' : ''} ` : '';
    const time = (interval.months > 0 || interval.days > 1) ? '' : microsToTimestring(interval.micros);
    // if only days && days > 365, convert to years?
    if (interval.months === 0 && interval.days > 0 && interval.days > 365) return `${formatRate(interval.days / 365)} years`
    return `${months}${days}${time}`;
}

export function formatCardinality(n:number) {
    let fmt:Function;
    if (n <= 1000) {
        fmt = formatInteger;
        return fmt(~~n);
    } else {
        fmt = format('.3s');
        return fmt(n);
    }
}