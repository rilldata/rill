import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DateTime } from "luxon";
import type { DateObjectUnits } from "luxon/src/datetime";
import {
  getMaxGrain,
  getMinGrain,
  grainAliasToDateTimeUnit,
  GrainAliasToV1TimeGrain,
} from "@rilldata/web-common/lib/time/new-grains";

const absTimeRegex =
  /(?<year>\d{4})(-(?<month>\d{2})(-(?<day>\d{2})(T(?<hour>\d{2})(:(?<minute>\d{2})(:(?<second>\d{2})Z)?)?)?)?)?/;
const simplifiedSnapMap: Record<V1TimeGrain, V1TimeGrain> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: V1TimeGrain.TIME_GRAIN_MILLISECOND,
  [V1TimeGrain.TIME_GRAIN_SECOND]: V1TimeGrain.TIME_GRAIN_SECOND,
  [V1TimeGrain.TIME_GRAIN_MINUTE]: V1TimeGrain.TIME_GRAIN_MINUTE,
  [V1TimeGrain.TIME_GRAIN_HOUR]: V1TimeGrain.TIME_GRAIN_HOUR,
  [V1TimeGrain.TIME_GRAIN_DAY]: V1TimeGrain.TIME_GRAIN_HOUR,
  [V1TimeGrain.TIME_GRAIN_WEEK]: V1TimeGrain.TIME_GRAIN_DAY,
  [V1TimeGrain.TIME_GRAIN_MONTH]: V1TimeGrain.TIME_GRAIN_DAY,
  [V1TimeGrain.TIME_GRAIN_QUARTER]: V1TimeGrain.TIME_GRAIN_DAY,
  [V1TimeGrain.TIME_GRAIN_YEAR]: V1TimeGrain.TIME_GRAIN_DAY,
};

export class RillTime {
  public timeRange: string;
  public readonly isComplete: boolean = false;
  public timezone: string | undefined;

  public readonly rangeGrain: V1TimeGrain | undefined;
  public readonly inGrain: V1TimeGrain | undefined;
  public byGrain: V1TimeGrain | undefined;
  public readonly isShorthandSyntax: boolean;

  public constructor(public readonly interval: RillTimeInterval) {
    this.isComplete = !this.interval.includesFuture;

    this.isShorthandSyntax =
      interval instanceof RillShorthandInterval ||
      interval instanceof RillPeriodToGrainInterval;
    [this.rangeGrain, this.inGrain] = this.interval.getGrains();
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
    return capitalizeFirstChar(supported ? label : this.timeRange);
  }

  public getCorrectInGrain(smallestTimeGrain: V1TimeGrain | undefined) {
    if (!this.inGrain) return undefined;
    if (!smallestTimeGrain) return this.inGrain;
    return getMaxGrain(this.inGrain, smallestTimeGrain);
  }

  public toString() {
    return this.timeRange;
  }
}

interface RillTimeInterval {
  includesFuture: boolean;

  getLabel(): [label: string, supported: boolean];
  getGrains(): [
    rangeGrain: V1TimeGrain | undefined,
    inGrain: V1TimeGrain | undefined,
  ];
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

  public getGrains(): [
    rangeGrain: V1TimeGrain | undefined,
    inGrain: V1TimeGrain | undefined,
  ] {
    let rangeGrain: V1TimeGrain | undefined = this.point?.getGrain();

    this.grains.forEach((grain) => {
      rangeGrain = getMinGrain(
        rangeGrain,
        GrainAliasToV1TimeGrain[grain.grain],
      );
    });

    return [rangeGrain, undefined];
  }
}

export class RillShorthandInterval implements RillTimeInterval {
  public constructor(
    private readonly num: number,
    private readonly grain: string,
    private readonly inGrain: string | undefined,
    public readonly includesFuture: boolean,
  ) {}

  public getLabel(): [label: string, supported: boolean] {
    const grainPart = grainAliasToDateTimeUnit(this.grain as any);

    if (this.num === 1) {
      return [
        `${this.includesFuture ? "this" : "previous"} ${grainPart}`,
        true,
      ];
    }

    const grainSuffix = this.num > 1 ? "s" : "";
    const grainLabel = `${this.num} ${grainPart}${grainSuffix}`;

    return [`last ${grainLabel}`, true];
  }

  public getGrains(): [
    rangeGrain: V1TimeGrain | undefined,
    inGrain: V1TimeGrain | undefined,
  ] {
    const rangeGrain = GrainAliasToV1TimeGrain[this.grain];
    console.log("INGRAIN", this.inGrain);
    return [
      rangeGrain,

      this.inGrain
        ? GrainAliasToV1TimeGrain[this.inGrain]
        : simplifiedSnapMap[rangeGrain],
    ];
  }
}

export class RillPeriodToGrainInterval implements RillTimeInterval {
  public constructor(
    private readonly grain: string,
    private readonly inGrain: string | undefined,
    public readonly includesFuture: boolean,
  ) {}

  public getLabel(): [label: string, supported: boolean] {
    return ["", false];
  }

  public getGrains(): [
    rangeGrain: V1TimeGrain | undefined,
    inGrain: V1TimeGrain | undefined,
  ] {
    const rangeGrain = GrainAliasToV1TimeGrain[this.grain];
    return [
      rangeGrain,
      simplifiedSnapMap[
        this.inGrain ? GrainAliasToV1TimeGrain[this.inGrain] : rangeGrain
      ],
    ];
  }
}

export class RillTimeOrdinalInterval implements RillTimeInterval {
  public includesFuture = false; // TODO: anything snapped to end before current should be true here

  public constructor(
    private readonly parts: RillOrdinalPart[],
    private readonly end: RillOrdinalPartEnd | undefined,
  ) {}

  public getLabel(): [label: string, supported: boolean] {
    return ["", false];
  }

  public getGrains(): [
    rangeGrain: V1TimeGrain | undefined,
    inGrain: V1TimeGrain | undefined,
  ] {
    let rangeGrain: V1TimeGrain | undefined = undefined;

    this.parts.forEach((part) => {
      rangeGrain = getMinGrain(rangeGrain, GrainAliasToV1TimeGrain[part.grain]);
    });

    if (this.end) {
      rangeGrain = getMinGrain(rangeGrain, this.end.getGrain());
    }

    return [rangeGrain, undefined];
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
      if (
        this.start instanceof RillGrainPointInTime &&
        typeof this.end === "string"
      ) {
        if (this.end === "latest" || this.end === "watermark") {
          const start = this.start.getSingleGrainAndNum();
          if (!start) return ["", false];
          const numDiff = Math.abs(start.offset);

          const grainPart = grainAliasToDateTimeUnit(start.grain as any);
          const grainSuffix = numDiff > 1 ? "s" : "";
          const grainPrefix = numDiff ? numDiff + " " : "";
          const grainLabel = `${grainPrefix}${grainPart}${grainSuffix}`;

          return [`last ${grainLabel}`, true];
        }
      }

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

  public getGrains(): [
    rangeGrain: V1TimeGrain | undefined,
    inGrain: V1TimeGrain | undefined,
  ] {
    const startRangeGrain = this.start.getGrain();
    const endRangeGrain =
      typeof this.end?.getGrain === "function"
        ? this.end.getGrain()
        : "TIME_GRAIN_DAY";
    const rangeGrain = getMinGrain(startRangeGrain, endRangeGrain);
    return [rangeGrain, undefined];
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

  public getGrains(): [
    rangeGrain: V1TimeGrain | undefined,
    inGrain: V1TimeGrain | undefined,
  ] {
    return [this.point.getGrain(), undefined];
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

  public getGrains(): [
    rangeGrain: V1TimeGrain | undefined,
    inGrain: V1TimeGrain | undefined,
  ] {
    return [undefined, undefined];
  }
}

interface RillPointInTime {
  includesFuture: boolean;
  getGrain(): V1TimeGrain | undefined;
}

export class RillOrdinalPointInTime implements RillPointInTime {
  public includesFuture = false;

  private restOfParts: RillOrdinalPart[] = [];
  private end: RillOrdinalPartEnd | undefined = undefined;
  private readonly ceil: boolean;

  public constructor(
    private readonly ordinal: RillOrdinalPart,
    suffix: string,
  ) {
    this.ceil = suffix === "$";
  }

  public withRestOfParts(restOfParts: RillOrdinalPart[]) {
    this.restOfParts = restOfParts;
    return this;
  }

  public withEnd(end: RillOrdinalPartEnd) {
    this.end = end;
    return this;
  }

  public getGrain(): V1TimeGrain | undefined {
    let rangeGrain = GrainAliasToV1TimeGrain[this.ordinal.grain] as
      | V1TimeGrain
      | undefined;

    this.restOfParts.forEach((part) => {
      rangeGrain = getMinGrain(rangeGrain, GrainAliasToV1TimeGrain[part.grain]);
    });

    if (this.end) {
      rangeGrain = getMinGrain(rangeGrain, this.end.getGrain());
    }

    return rangeGrain;
  }
}

export class RillOrdinalPart {
  public constructor(
    public readonly grain: string,
    public readonly num: number | undefined,
    public readonly snap: string | undefined,
  ) {}
}

export class RillOrdinalPartEnd {
  private grainToInterval: RillGrainToInterval | undefined = undefined;
  private startEndInterval: RillTimeStartEndInterval | undefined = undefined;
  private singleGrain: string | undefined = undefined;

  public withGrainToInterval(grainToInterval: RillGrainToInterval) {
    this.grainToInterval = grainToInterval;
    return this;
  }

  public withStartEndInterval(startEndInterval: RillTimeStartEndInterval) {
    this.startEndInterval = startEndInterval;
    return this;
  }

  public withSingleGrain(singleGrain: string) {
    this.singleGrain = singleGrain;
    return this;
  }

  public getGrain(): V1TimeGrain | undefined {
    let rangeGrain: V1TimeGrain | undefined = undefined;

    if (this.grainToInterval) {
      [rangeGrain] = this.grainToInterval.getGrains();
    } else if (this.startEndInterval) {
      [rangeGrain] = this.startEndInterval.getGrains();
    } else if (this.singleGrain) {
      rangeGrain = GrainAliasToV1TimeGrain[this.singleGrain];
    }

    return rangeGrain;
  }
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
