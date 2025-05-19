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
    this.isComplete = !this.interval.includesFuture;
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
  includesFuture: boolean;

  getLabel(): [label: string, supported: boolean];
}

export class RillTimeAnchoredDurationInterval implements RillTimeInterval {
  public includesFuture: boolean;

  public constructor(
    public readonly grains: RillGrain[],
    public readonly starting: boolean,
    public readonly point: RillPointInTime,
  ) {
    // If this ends before current, then it is guaranteed to be complete.
    const endingBeforeCurrent = !starting && !point.includesFuture;
    this.includesFuture = !endingBeforeCurrent;
  }

  public getLabel(): [label: string, supported: boolean] {
    return ["", false];
  }
}

export class RillTimeOrdinalInterval implements RillTimeInterval {
  public includesFuture = false; // TODO: anything snapped to end before current should be true here

  public getLabel(): [label: string, supported: boolean] {
    return ["", false];
  }
}

export class RillTimeStartEndInterval implements RillTimeInterval {
  public includesFuture = false;

  public constructor(
    public readonly start: RillPointInTime,
    public readonly end: RillPointInTime,
  ) {
    this.includesFuture = start.includesFuture || end.includesFuture;
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

    const numDiff = Math.abs(start.offset - end.offset);
    if (start.grain !== end.grain) {
      if (numDiff > 1) {
        return ["", false];
      }

      const startLabel = grainAliasToDateTimeUnit(start.grain as any);
      const endLabel = grainAliasToDateTimeUnit(end.grain as any);
      return [`${startLabel} to ${endLabel}`, true];
    }

    const grainPart = grainAliasToDateTimeUnit(start.grain as any);
    const grainSuffix = numDiff > 1 ? "s" : "";
    const grainPrefix = numDiff ? numDiff + " " : "";
    const grainLabel = `${grainPrefix}${grainPart}${grainSuffix}`;

    if (start.offset === 0 || start.offset === 1) {
      if (numDiff === 1) {
        const prefix = start.offset === 0 ? "this" : "next";
        return [`${prefix} ${grainPart}`, true];
      }
      return [`next ${grainLabel}`, true];
    }

    if (end.offset === 0 || end.offset === 1) {
      if (numDiff === 1) {
        const prefix = end.offset === 1 ? "this" : "previous";
        return [`${prefix} ${grainPart}`, true];
      }
      return [`last ${grainLabel}`, true];
    }

    return ["", false];
  }
}

export class RillGrainToInterval implements RillTimeInterval {
  public includesFuture = false;

  public constructor(public readonly point: RillGrainPointInTime) {
    const first = point.getSingleGrainAndNum();
    if (!first) return;
    this.includesFuture = first.offset >= 0 || first.offset === undefined;
  }

  public getLabel(): [label: string, supported: boolean] {
    const grainAndNum = this.point.getSingleGrainAndNum();
    if (!grainAndNum) return ["", false];

    const label = grainAliasToDateTimeUnit(grainAndNum.grain as any);

    if (grainAndNum.offset === 0) {
      return [`this ${label}`, true];
    } else if (grainAndNum.offset === 1) {
      return [`next ${label}`, true];
    } else if (grainAndNum.offset === -1) {
      return [`previous ${label}`, true];
    } else {
      return ["", true];
    }
  }
}

export class RillIsoInterval implements RillTimeInterval {
  public includesFuture = false;

  public getLabel(): [label: string, supported: boolean] {
    return ["", false];
  }
}

interface RillPointInTime {
  includesFuture: boolean;
}

export class RillOrdinalPointInTime implements RillPointInTime {
  public includesFuture = false;
}

export class RillGrainPointInTime implements RillPointInTime {
  public includesFuture: boolean;

  public constructor(public readonly parts: RillGrainPointInTimePart[]) {
    this.includesFuture = parts[0]?.includesFuture ?? false;
  }

  public getSingleGrainAndNum() {
    if (this.parts.length !== 1) return undefined;
    const firstPart = this.parts[0];
    if (firstPart.grains.length !== 1) return undefined;
    const firstGrain = firstPart.grains[0];

    let offset = firstGrain.num ?? 0;
    if (firstPart.prefix === "-" && offset) {
      // Grain doesn't have a `-` inbuilt, make the offset negative.
      offset = -offset;
    }
    if (firstPart.suffix === "$") {
      // Since xx$ will snap to the end, so add 1 to the offset
      offset++;
    }

    return {
      grain: firstGrain.grain,
      offset,
      firstPart,
      firstGrain,
    };
  }
}
export class RillGrainPointInTimePart {
  public prefix: string;
  public snap: string;
  public suffix: string;

  public includesFuture = false;

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

  private updateAfterCurrent() {
    const firstGrain = this.grains[0];
    if (!firstGrain) return;

    const firstGrainNum = firstGrain.num ?? 0;

    if (firstGrainNum === 0 && this.suffix !== undefined) {
      this.includesFuture = this.suffix === "$";
    } else {
      this.includesFuture = this.prefix === "+";
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

type RillGrain = {
  grain: string;
  num?: number;
};

function capitalizeFirstChar(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}
