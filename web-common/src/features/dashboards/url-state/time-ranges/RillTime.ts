import { DateTime } from "luxon";
import type { DateObjectUnits } from "luxon/src/datetime";
import { grainAliasToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";

const absTimeRegex =
  /(?<year>\d{4})(-(?<month>\d{2})(-(?<day>\d{2})(T(?<hour>\d{2})(:(?<minute>\d{2})(:(?<second>\d{2})Z)?)?)?)?)?/;

export class RillTime {
  public timeRange: string;
  public readonly isComplete: boolean = false;
  public timeRangeGrain: string | undefined;
  public timezone: string | undefined;

  public constructor(public readonly interval: RillTimeInterval) {
    this.isComplete = !this.interval.includesCurrent;
  }

  public withGrain(grain: string) {
    this.timeRangeGrain = grain;
    return this;
  }

  public withTimeZone(timezone: string) {
    this.timezone = timezone;
    return this;
  }

  public getLabel() {
    console.log("GETTING LABEL");
    const [label, supported] = this.interval.getLabel();
    return capitalizeFirstChar(supported ? label : this.timeRange);
  }

  public toString() {
    return this.timeRange;
  }
}

interface RillTimeInterval {
  includesCurrent: boolean;

  getLabel(): [label: string, supported: boolean];
}

export class RillTimeAnchoredDurationInterval implements RillTimeInterval {
  public includesCurrent: boolean;

  public constructor(
    public readonly grains: RillGrain[],
    public readonly starting: boolean,
    public readonly point: RillPointInTime,
  ) {
    // If this ends before current, then it is guaranteed to be complete.
    const endingBeforeCurrent = !starting && !point.afterCurrent;
    this.includesCurrent = !endingBeforeCurrent;
  }

  public getLabel(): [label: string, supported: boolean] {
    return ["", false];
  }
}

export class RillTimeOrdinalInterval implements RillTimeInterval {
  public includesCurrent = false; // TODO: anything snapped to end before current should be true here

  public getLabel(): [label: string, supported: boolean] {
    return ["", false];
  }
}

export class RillTimeStartEndInterval implements RillTimeInterval {
  public includesCurrent = false;

  public constructor(
    public readonly start: RillPointInTime,
    public readonly end: RillPointInTime,
  ) {
    this.includesCurrent = !start.afterCurrent && end.afterCurrent;
  }

  public getLabel(): [label: string, supported: boolean] {
    if (
      !(this.start instanceof RillGrainPointInTime) ||
      !(this.end instanceof RillGrainPointInTime)
    ) {
      return ["", false];
    }

    const start = this.start.getSingleGrainAndNum();
    const end = this.end.getSingleGrainAndNum();
    if (!start || !end) return ["", false];

    const numDiff = start.firstPart.diff(start.num, end.firstPart, end.num);
    if (start.grain !== end.grain) {
      if (numDiff > 1) {
        return ["", false];
      }

      const startLabel = grainAliasToDateTimeUnit(start.grain as any);
      const endLabel = grainAliasToDateTimeUnit(end.grain as any);
      return [`${startLabel} to ${endLabel}`, true];
    }

    const grainPart = grainAliasToDateTimeUnit(start.grain);
    const grainSuffix = numDiff > 1 ? "s" : "";
    const grainPrefix = numDiff ? numDiff + " " : "";
    const grainLabel = `${grainPrefix}${grainPart}${grainSuffix}`;

    if (!start.firstPart.afterCurrent && !end.firstPart.afterCurrent) {
      if (end.num < -1) {
        return ["", false];
      }
      if (numDiff === 1) {
        return [`previous ${grainPart}`, true];
      }
      return [`last ${grainLabel}`, true];
    }

    if (start.firstPart.afterCurrent && end.firstPart.afterCurrent) {
      if (end.num > 1) {
        return ["", false];
      }
      return [`next ${grainLabel}`, true];
    }

    if (numDiff === 1) return [`this ${grainPart}`, true];

    return ["", false];
  }
}

export class RillGrainToInterval implements RillTimeInterval {
  public includesCurrent = false;

  public constructor(public readonly point: RillGrainPointInTime) {
    const firstGrainOfFirstPart = point.parts[0]?.grains[0];
    if (!firstGrainOfFirstPart) return;
    this.includesCurrent =
      firstGrainOfFirstPart.num === 0 ||
      firstGrainOfFirstPart.num === undefined;
  }

  public getLabel(): [label: string, supported: boolean] {
    const grainAndNum = this.point.getSingleGrainAndNum();
    if (!grainAndNum) return ["", false];

    const label = grainAliasToDateTimeUnit(grainAndNum.grain as any);

    if (grainAndNum.num === 0) {
      return [`this ${label}`, true];
    } else if (grainAndNum.num === 1) {
      return [`next ${label}`, true];
    } else if (grainAndNum.num === -1) {
      return [`previous ${label}`, true];
    } else {
      return ["", true];
    }
  }
}

export class RillIsoInterval implements RillTimeInterval {
  public includesCurrent = false;

  public getLabel(): [label: string, supported: boolean] {
    return ["", false];
  }
}

interface RillPointInTime {
  afterCurrent: boolean;
}

export class RillOrdinalPointInTime implements RillPointInTime {
  public afterCurrent = false;
}

export class RillGrainPointInTime implements RillPointInTime {
  public afterCurrent: boolean;

  public constructor(public readonly parts: RillGrainPointInTimePart[]) {
    this.afterCurrent = parts[0]?.afterCurrent ?? false;
  }

  public getSingleGrainAndNum() {
    if (this.parts.length !== 1) return undefined;
    const firstPart = this.parts[0];
    if (firstPart.grains.length !== 1) return undefined;
    const firstGrain = firstPart.grains[0];

    let num = firstGrain.num ?? 0;
    if (firstPart.prefix === "-" && num) {
      num = -num;
    }

    return {
      grain: firstGrain.grain,
      num,
      firstPart,
      firstGrain,
    };
  }
}
export class RillGrainPointInTimePart {
  public prefix: string;
  public snap: string;
  public suffix: string;

  public afterCurrent = false;

  public constructor(public readonly grains: RillGrain[]) {
    this.updateAfterCurrent();
  }

  public withPrefix(prefix: string) {
    this.prefix = prefix;
    this.updateAfterCurrent();
    return this;
  }

  public withSnap(snap: string) {
    this.snap = snap;
    this.updateAfterCurrent();
    return this;
  }

  public withSuffix(suffix: string) {
    this.suffix = suffix;
    this.updateAfterCurrent();
    return this;
  }

  public diff(num1: number, part2: RillGrainPointInTimePart, num2: number) {
    const offset = this.suffix === part2.suffix ? 0 : 1;
    return Math.abs(num1 - num2) + offset;
  }

  private updateAfterCurrent() {
    const firstGrain = this.grains[0];
    if (!firstGrain) return;

    const firstGrainNum = firstGrain.num ?? 0;

    if (firstGrainNum === 0 && this.suffix !== undefined) {
      this.afterCurrent = this.suffix === "$";
    } else {
      this.afterCurrent = this.prefix === "+";
    }
  }
}

interface RillTimePart {
  getLabel(): string;
  toString(): string;
  isComplete: boolean;
}

export class RillAbsoluteTime implements RillTimePart {
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
    return new RillAbsoluteTime(args.join(""));
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
    const grainPart = grainAliasToDateTimeUnit(this.grain);
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
    const grainPart = grainAliasToDateTimeUnit(this.grain);
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

type RillGrain = {
  grain: string;
  num?: number;
};

function capitalizeFirstChar(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}
