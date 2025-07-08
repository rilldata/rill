import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DateTime } from "luxon";
import type { DateObjectUnits } from "luxon/src/datetime";
import {
  getMinGrain,
  grainAliasToDateTimeUnit,
  GrainAliasToV1TimeGrain,
} from "@rilldata/web-common/lib/time/new-grains";

const absTimeRegex =
  /(?<year>\d{4})(-(?<month>\d{2})(-(?<day>\d{2})(T(?<hour>\d{2})(:(?<minute>\d{2})(:(?<second>\d{2})Z)?)?)?)?)?/;

export class RillTime {
  public timeRange: string;
  public readonly isComplete: boolean = false;
  public timezone: string | undefined;

  public readonly rangeGrain: V1TimeGrain | undefined;
  public byGrain: V1TimeGrain | undefined;
  public readonly isShorthandSyntax: boolean;

  public constructor(public readonly interval: RillTimeInterval) {
    this.isComplete = !this.interval.includesFuture;

    this.isShorthandSyntax =
      interval instanceof RillShorthandInterval ||
      interval instanceof RillPeriodToGrainInterval;
    this.rangeGrain = this.interval.getGrains();
  }

  public withGrain(grain: string) {
    this.byGrain = GrainAliasToV1TimeGrain[grain];
    return this;
  }

  public withTimezone(timezone: string) {
    this.timezone = timezone;
    return this;
  }

  public getLabel() {
    const [label, supported] = this.interval.getLabel();
    return supported ? capitalizeFirstChar(label) : this.timeRange;
  }

  public toString() {
    return this.timeRange;
  }
}

interface RillTimeInterval {
  includesFuture: boolean;

  getLabel(): [label: string, supported: boolean];
  getGrains(): V1TimeGrain | undefined;
}

export class RillShorthandInterval implements RillTimeInterval {
  public includesFuture: boolean;

  private readonly expandedInterval: RillTimeStartEndInterval;

  public constructor(parts: RillGrainPointInTimePart[]) {
    this.expandedInterval = new RillTimeStartEndInterval(
      new RillPointInTime([
        new RillPointInTimeWithSnap(new RillGrainPointInTime(parts), []),
      ]),
      new RillPointInTime([
        new RillPointInTimeWithSnap(new RillLabelledPointInTime("ref"), []),
      ]),
    );
    this.includesFuture = this.expandedInterval.includesFuture;
  }

  public getLabel(): [label: string, supported: boolean] {
    return this.expandedInterval.getLabel();
  }

  public getGrains() {
    return this.expandedInterval.getGrains();
  }
}

export class RillPeriodToGrainInterval implements RillTimeInterval {
  public includesFuture = true;

  private readonly expandedInterval: RillTimeStartEndInterval;

  public constructor(grain: string) {
    this.expandedInterval = new RillTimeStartEndInterval(
      new RillPointInTime([
        new RillPointInTimeWithSnap(new RillLabelledPointInTime("ref"), [
          grain,
        ]),
      ]),
      new RillPointInTime([
        new RillPointInTimeWithSnap(new RillLabelledPointInTime("ref"), []),
      ]),
    );
  }

  public getLabel(): [label: string, supported: boolean] {
    return this.expandedInterval.getLabel();
  }

  public getGrains() {
    return this.expandedInterval.getGrains();
  }
}

export class RillTimeOrdinalInterval implements RillTimeInterval {
  public includesFuture = false; // TODO: anything snapped to end before current should be true here

  public constructor(private readonly parts: RillOrdinal[]) {}

  public getLabel(): [label: string, supported: boolean] {
    return ["", false];
  }

  public getGrains() {
    let rangeGrain: V1TimeGrain | undefined = undefined;

    this.parts.forEach((part) => {
      rangeGrain = getMinGrain(rangeGrain, GrainAliasToV1TimeGrain[part.grain]);
    });

    return rangeGrain;
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
    const start = this.start.getSingleGrainAndNum();
    const end = this.end.getSingleGrainAndNum();
    if (!start || !end) return ["", false];

    if (start.isLabelled || end.isLabelled) {
      if (!start.isLabelled) {
        const numDiff = Math.abs(start.offset);

        const grainPart = grainAliasToDateTimeUnit(start.grain as any);
        const grainSuffix = numDiff > 1 ? "s" : "";
        const grainPrefix = numDiff ? numDiff + " " : "";
        const grainLabel = `${grainPrefix}${grainPart}${grainSuffix}`;

        return [`last ${grainLabel}`, true];
      }

      return ["", false];
    }

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

  public getGrains() {
    const startRangeGrain = this.start.getGrain();
    const endRangeGrain =
      typeof this.end?.getGrain === "function"
        ? this.end.getGrain()
        : "TIME_GRAIN_DAY";
    const rangeGrain = getMinGrain(startRangeGrain, endRangeGrain);
    return rangeGrain;
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

  public getGrains() {
    return this.point.getGrain();
  }
}

export class RillIsoInterval implements RillTimeInterval {
  public includesFuture = false;

  public constructor(
    private readonly start: RillAbsoluteTime,
    private readonly end: RillAbsoluteTime | undefined,
  ) {}

  public getLabel(): [label: string, supported: boolean] {
    return ["", false];
  }

  public getGrains() {
    return undefined;
  }
}

type SingleGrainAndNum =
  | {
      isLabelled: true;
    }
  | {
      isLabelled: false;
      grain: string;
      offset: number;
      firstPart: RillGrainPointInTimePart;
      firstGrain: RillGrain;
    };

export class RillPointInTime {
  public includesFuture = false;

  public constructor(public readonly parts: RillPointInTimeWithSnap[]) {
    this.includesFuture = parts[0]?.point.includesFuture ?? false;
  }

  public getSingleGrainAndNum(): SingleGrainAndNum | undefined {
    if (this.parts.length !== 1) return undefined;
    const singlePart = this.parts[0];
    if (singlePart.point instanceof RillGrainPointInTime) {
      return singlePart.point.getSingleGrainAndNum();
    } else if (singlePart.point instanceof RillLabelledPointInTime) {
      return {
        isLabelled: true,
      };
    }
    return undefined;
  }

  public getGrain(): V1TimeGrain | undefined {
    let rangeGrain: V1TimeGrain | undefined = undefined;
    this.parts.forEach((part) => {
      rangeGrain = getMinGrain(rangeGrain, part.point.getGrain());
    });
    return rangeGrain;
  }
}

export class RillPointInTimeWithSnap {
  public constructor(
    public readonly point: RillPointInTimeVariant,
    private snaps: string[],
  ) {}
}

interface RillPointInTimeVariant {
  includesFuture: boolean;
  getGrain(): V1TimeGrain | undefined;
}

export type RillOrdinal = {
  grain: string;
  num: number;
};

export class RillGrainPointInTime implements RillPointInTimeVariant {
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

    return {
      isLabelled: false,
      grain: firstGrain.grain,
      offset,
      firstPart,
      firstGrain,
    };
  }

  public getGrain(): V1TimeGrain | undefined {
    let rangeGrain: V1TimeGrain | undefined = undefined;

    this.parts.forEach((part) => {
      part.grains.forEach((grain) => {
        rangeGrain = getMinGrain(
          rangeGrain,
          GrainAliasToV1TimeGrain[grain.grain],
        );
      });
    });

    return rangeGrain;
  }
}

export class RillGrainPointInTimePart {
  public includesFuture = false;

  public constructor(
    public readonly prefix: string,
    public readonly grains: RillGrain[],
  ) {
    const firstGrain = this.grains[0];
    if (!firstGrain) return;

    const firstGrainNum = firstGrain.num ?? 0;

    this.includesFuture = firstGrainNum > 0 && this.prefix === "+";
  }
}

export class RillLabelledPointInTime implements RillPointInTimeVariant {
  public includesFuture = false;

  public constructor(private readonly label: string) {}

  public static postProcessor([label]: string[]) {
    return new RillLabelledPointInTime(label);
  }

  public getGrain(): V1TimeGrain | undefined {
    return undefined;
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
