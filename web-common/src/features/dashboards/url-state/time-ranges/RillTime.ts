import { DateTime } from "luxon";
import type { DateObjectUnits } from "luxon/src/datetime";

const absTimeRegex =
  /(?<year>\d{4})(-(?<month>\d{2})(-(?<day>\d{2})(T(?<hour>\d{2})(:(?<minute>\d{2})(:(?<second>\d{2})Z)?)?)?)?)?/;

export class RillTime {
  public timeRange: string;
  public readonly isComplete: boolean;

  public constructor(
    public readonly start: RillTimePart[],
    public readonly end: RillTimePart[] | undefined,
    public readonly timeRangeGrain: string | undefined,
    public timezone: string | undefined,
  ) {
    this.isComplete = end?.[0]?.isComplete ?? start[0]?.isComplete ?? false;
  }

  public withTimeZone(timezone: string) {
    this.timezone = timezone;
    return this;
  }

  public getLabel() {
    if (this.end) return this.timeRange; // TODO: what would the labels be here?

    let range = this.start.map((p) => p.getLabel()).join(" of ");

    if (this.timezone) {
      range += ` in ${this.timezone}`;
    }

    return capitalizeFirstChar(range);
  }

  public toString() {
    let range = this.start.map((p) => p.toString()).join(" of ");

    if (this.end) {
      range += ` to ${this.end.map((p) => p.toString()).join(" of ")}`;
    }

    if (this.timeRangeGrain) {
      range += ` by ${this.timeRangeGrain}`;
    }

    if (this.timezone) {
      range += ` tz ${this.timezone}`;
    }

    return range;
  }
}

const GrainToUnit = {
  s: "second",
  S: "second",
  m: "minute",
  h: "hour",
  H: "hour",
  d: "day",
  D: "day",
  w: "week",
  W: "week",
  M: "month",
  q: "Quarter",
  Q: "Quarter",
  y: "year",
  Y: "year",
};

interface RillTimePart {
  getLabel(): string;
  toString(): string;
  isComplete: boolean;
}

export class RillTimeAbsoluteTime implements RillTimePart {
  public isComplete = true; // TODO: can this be anything else?

  private readonly dateObject: DateObjectUnits = {};

  public constructor(private readonly timeStr: string) {
    const absTimeMatch = absTimeRegex.exec(timeStr);
    if (!absTimeMatch) {
      return;
    }

    if (absTimeMatch.groups?.year)
      this.dateObject.year = Number(absTimeMatch.groups.year);
    if (absTimeMatch.groups?.month)
      this.dateObject.month = Number(absTimeMatch.groups.month);
    if (absTimeMatch.groups?.day)
      this.dateObject.day = Number(absTimeMatch.groups.day);
    if (absTimeMatch.groups?.hour)
      this.dateObject.hour = Number(absTimeMatch.groups.hour);
    if (absTimeMatch.groups?.minute)
      this.dateObject.minute = Number(absTimeMatch.groups.minute);
    if (absTimeMatch.groups?.second)
      this.dateObject.second = Number(absTimeMatch.groups.second);
  }

  public static postProcessor(args: string[]) {
    return new RillTimeAbsoluteTime(args.join(""));
  }

  public getLabel() {
    const date = DateTime.fromObject(this.dateObject, { zone: "utc" });

    if (
      this.dateObject.hour ||
      this.dateObject.minute ||
      this.dateObject.second
    ) {
      return date.toLocaleString(DateTime.DATETIME_MED);
    }

    if (this.dateObject.day) {
      return date.toLocaleString(DateTime.DATE_MED);
    }

    if (this.dateObject.month) {
      return date.toLocaleString({ month: "short", year: "numeric" });
    }

    return this.timeStr;
  }

  public toString() {
    return this.timeStr;
  }
}

export class RillTimeLabelledAnchor implements RillTimePart {
  public isComplete = false; // TODO: can this be anything else?

  public constructor(public readonly label: string) {}

  public static postProcessor([label]: [string]) {
    return new RillTimeLabelledAnchor(label);
  }

  public getLabel() {
    return this.label;
  }

  public toString() {
    return this.label;
  }
}

export class RillTimeOrdinal implements RillTimePart {
  public isComplete = true;

  public constructor(
    private readonly grain: string,
    private readonly num: number,
  ) {}

  public getLabel() {
    const grainPart = GrainToUnit[this.grain];
    return `${grainPart} ${this.num}`;
  }

  public toString() {
    return `${this.grain}${this.num}`;
  }
}

export class RillTimeRelative implements RillTimePart {
  public isComplete = true;

  public constructor(
    private readonly prefix: "+" | "-" | "<" | ">" | undefined,
    private readonly num: number,
    private readonly grain: string,
  ) {}

  public asIncomplete() {
    this.isComplete = false;
    return this;
  }

  public getLabel() {
    const grainPart = GrainToUnit[this.grain];
    const grainSuffix = this.num > 1 ? "s" : "";
    const grainPrefix = this.num ? this.num + " " : "";
    const grainLabel = `${grainPrefix}${grainPart}${grainSuffix}`;

    switch (this.prefix) {
      case undefined:
        if (this.num === 1) {
          return `${this.isComplete ? "previous" : "this"} ${grainPart}`;
        }
        return `last ${grainLabel}`;

      case "-":
        if (this.num === 1) {
          return `previous ${grainPart}`;
        }
        return `${grainLabel} ago`;

      case "+":
        if (this.num === 1) {
          return `next ${grainPart}`;
        }
        return `${grainLabel} in the future`;

      case "<":
        return `first ${grainLabel}`;

      case ">":
        return `last ${grainLabel}`;
    }
  }

  public toString() {
    return (
      `${this.prefix ?? ""}${this.num != 0 ? this.num : ""}` +
      `${this.grain}${this.isComplete ? "" : "~"}`
    );
  }
}

export class RillTimePeriodToDate implements RillTimePart {
  private readonly from: string;
  private readonly to: string;
  public isComplete = true;

  public constructor(
    private readonly prefix: "+" | "-" | undefined,
    private readonly num: number,
    private readonly periodToDate: string,
  ) {
    [this.from, this.to] = periodToDate.split("T");
  }

  public asIncomplete() {
    this.isComplete = false;
    return this;
  }

  public getLabel() {
    const from = GrainToUnit[this.from];
    const to = GrainToUnit[this.to];
    // TODO
    return `${from} to ${to}`;
  }

  public toString() {
    return (
      `${this.prefix ?? ""}${this.num != 0 ? this.num : ""}` +
      `${this.periodToDate}${this.isComplete ? "" : "~"}`
    );
  }
}

function capitalizeFirstChar(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}
