import { format } from "d3-format";

const zeroPad = format('02d');
const formatInteger = format(',');

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
    const time = microsToTimestring(interval.micros);
    return `${months}${days}${time}`;
}