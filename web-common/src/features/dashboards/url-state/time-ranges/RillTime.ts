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

export enum RillTimeLabel {
  Earliest = "earliest",
  Latest = "latest",
  Now = "now",
  Watermark = "watermark",
  Ref = "ref",
}

export class RillTime {
  public isComplete: boolean = false;
  public timezone: string | undefined;
  public anchorOverrides: RillPointInTime[] = [];

  public readonly rangeGrain: V1TimeGrain | undefined;
  public byGrain: V1TimeGrain | undefined;
  public readonly isShorthandSyntax: boolean;

  public constructor(public readonly interval: RillTimeInterval) {
    this.updateIsComplete();

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

  public withAnchorOverrides(anchorOverrides: RillPointInTime[]) {
    this.anchorOverrides = anchorOverrides;
    this.updateIsComplete();
    return this;
  }

  public getLabel() {
    const [offset, offsetSupported] = this.getAnchorOverridesOffset();
    if (!offsetSupported) return this.toString();

    const [label, supported] = this.interval.getLabel(offset);
    return supported ? capitalizeFirstChar(label) : this.toString();
  }

  public overrideRef(override: RillPointInTime) {
    const pointUsingRefIndex = this.anchorOverrides.findIndex((pt) =>
      pt.hasLabelledPart(),
    );
    if (pointUsingRefIndex >= 0) {
      this.anchorOverrides[pointUsingRefIndex] = override;
    } else {
      this.anchorOverrides.push(override);
    }
    this.updateIsComplete();
  }

  public toString() {
    let timeRange = this.interval.toString();

    timeRange += this.anchorOverrides
      .map((anchor) => ` AS OF ${anchor.toString()}`)
      .join("");

    if (this.byGrain) {
      timeRange += ` BY ${this.byGrain}`;
    }

    if (this.timezone) {
      timeRange += ` TZ ${this.timezone}`;
    }

    return timeRange;
  }

  private updateIsComplete() {
    const [offset, offsetSupported] = this.getAnchorOverridesOffset();
    if (!offsetSupported) this.isComplete = false;
    else this.isComplete = this.interval.isComplete(offset);
  }

  private getAnchorOverridesOffset(): [
    offset: RillGrainOffset | undefined,
    supported: boolean,
  ] {
    let offset: RillGrainOffset | undefined = undefined;
    let supported = true;

    this.anchorOverrides.forEach((anchor) => {
      const overrideOffset = anchor.getGrainOffset(offset);
      if (!overrideOffset) {
        supported = false;
        return;
      }
      offset = overrideOffset;
    });

    return [offset, supported];
  }
}

interface RillTimeInterval {
  isComplete(offset: RillGrainOffset | undefined): boolean;
  getLabel(
    offset: RillGrainOffset | undefined,
  ): [label: string, supported: boolean];
  getGrains(): V1TimeGrain | undefined;
  toString(): string;
}

export class RillShorthandInterval implements RillTimeInterval {
  private readonly expandedInterval: RillTimeStartEndInterval;

  public constructor(private readonly parts: RillGrain[]) {
    this.expandedInterval = new RillTimeStartEndInterval(
      new RillPointInTime([
        new RillPointInTimeWithSnap(
          new RillGrainPointInTime([new RillGrainPointInTimePart("-", parts)]),
          [],
        ),
      ]),
      new RillPointInTime([
        new RillPointInTimeWithSnap(
          new RillLabelledPointInTime(RillTimeLabel.Ref),
          [],
        ),
      ]),
    );
  }

  public isComplete(offset: RillGrainOffset | undefined) {
    return this.expandedInterval.isComplete(offset);
  }

  public getLabel(
    offset: RillGrainOffset | undefined,
  ): [label: string, supported: boolean] {
    return this.expandedInterval.getLabel(offset);
  }

  public getGrains() {
    return this.expandedInterval.getGrains();
  }

  public toString() {
    return this.parts
      .map((part) => {
        const grainPrefix = part.num ? part.num : "";
        return `${grainPrefix}${part.grain}`;
      })
      .join("");
  }
}

export class RillPeriodToGrainInterval implements RillTimeInterval {
  private readonly expandedInterval: RillTimeStartEndInterval;

  public constructor(private readonly grain: string) {
    this.expandedInterval = new RillTimeStartEndInterval(
      new RillPointInTime([
        new RillPointInTimeWithSnap(
          new RillLabelledPointInTime(RillTimeLabel.Ref),
          [grain],
        ),
      ]),
      new RillPointInTime([
        new RillPointInTimeWithSnap(
          new RillLabelledPointInTime(RillTimeLabel.Ref),
          [],
        ),
      ]),
    );
  }

  public isComplete(offset: RillGrainOffset | undefined) {
    return this.expandedInterval.isComplete(offset);
  }

  public getLabel(
    offset: RillGrainOffset | undefined,
  ): [label: string, supported: boolean] {
    return this.expandedInterval.getLabel(offset);
  }

  public getGrains() {
    return this.expandedInterval.getGrains();
  }

  public toString() {
    return `${this.grain}TD`;
  }
}

export class RillTimeOrdinalInterval implements RillTimeInterval {
  public constructor(private readonly parts: RillOrdinal[]) {}

  public isComplete() {
    return false;
  }

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

  public toString() {
    return this.parts.map((part) => `${part.grain}${part.num}`).join(" OF ");
  }
}

export class RillTimeStartEndInterval implements RillTimeInterval {
  public constructor(
    public readonly start: RillPointInTime,
    public readonly end: RillPointInTime,
  ) {}

  public isComplete(offset: RillGrainOffset | undefined) {
    const start = this.start.getGrainOffset(offset);
    const startIsComplete =
      start?.offset !== undefined ? start.offset <= 0 : true;
    const end = this.end.getGrainOffset(offset);
    const endIsComplete = end?.offset !== undefined ? end.offset <= 0 : true;
    return startIsComplete && endIsComplete;
  }

  public getLabel(
    offset: RillGrainOffset | undefined,
  ): [label: string, supported: boolean] {
    const start = this.start.getGrainOffset(offset);
    const end = this.end.getGrainOffset(offset);
    if (!start || !end) return ["", false];
    if (start.grain && end.grain && start.grain !== end.grain) {
      return ["", false];
    }

    const grain = start.grain || end.grain || "";
    const numDiff = Math.abs(start.offset - end.offset);

    const grainPart = grainAliasToDateTimeUnit(grain as any);
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

  public toString() {
    return `${this.start.toString()} TO ${this.end.toString()}`;
  }
}

export class RillIsoInterval implements RillTimeInterval {
  public constructor(
    private readonly start: RillAbsoluteTime,
    private readonly end: RillAbsoluteTime | undefined,
  ) {}

  public isComplete() {
    return false;
  }

  public getLabel(): [label: string, supported: boolean] {
    return ["", false];
  }

  public getGrains() {
    return undefined;
  }

  public toString() {
    let timeRange = this.start.toString();
    if (this.end) {
      timeRange += ` TO ${this.end.toString()}`;
    }
    return timeRange;
  }
}

export class RillPointInTime {
  public constructor(public readonly parts: RillPointInTimeWithSnap[]) {}

  public getGrainOffset(
    offset: RillGrainOffset | undefined,
  ): RillGrainOffset | undefined {
    let returnGrainOffset: RillGrainOffset | undefined = offset
      ? {
          ...offset,
        }
      : undefined;
    let notSupported = false;
    this.parts.forEach((part) => {
      const grainOffset = part.getGrainOffset();
      if (!grainOffset) {
        notSupported = true;
        return;
      }

      if (
        returnGrainOffset?.grain &&
        grainOffset.grain &&
        returnGrainOffset.grain !== grainOffset.grain
      ) {
        notSupported = true;
        return;
      }

      if (!returnGrainOffset) {
        returnGrainOffset = grainOffset;
      } else {
        returnGrainOffset.offset += grainOffset.offset;
      }
    });

    return notSupported ? undefined : returnGrainOffset;
  }

  public getGrain(): V1TimeGrain | undefined {
    let rangeGrain: V1TimeGrain | undefined = undefined;
    this.parts.forEach((part) => {
      rangeGrain = getMinGrain(rangeGrain, part.point.getGrain());
    });
    return rangeGrain;
  }

  public hasLabelledPart() {
    return this.parts.some((p) => p.point instanceof RillLabelledPointInTime);
  }

  public toString() {
    return this.parts.map((part) => part.toString()).join("");
  }
}

export class RillPointInTimeWithSnap {
  public constructor(
    public readonly point: RillPointInTimeVariant,
    private snaps: string[],
  ) {}

  public toString() {
    return `${this.point.toString()}${this.snaps.map((s) => "/" + s).join("")}`;
  }

  public getGrainOffset() {
    if (this.point instanceof RillGrainPointInTime) {
      const grainOffset = this.point.getGrainOffset();
      if (!grainOffset) return undefined;
      return this.snaps.every((s) => s === grainOffset.grain)
        ? grainOffset
        : undefined;
    } else if (this.point instanceof RillLabelledPointInTime) {
      return {
        grain: this.snaps[0],
        offset: 0,
      };
    } else {
      return undefined;
    }
  }
}

interface RillPointInTimeVariant {
  getGrain(): V1TimeGrain | undefined;
  toString(): string;
}

export type RillOrdinal = {
  grain: string;
  num: number;
};

export class RillGrainPointInTime implements RillPointInTimeVariant {
  public constructor(public readonly parts: RillGrainPointInTimePart[]) {}

  public getGrainOffset(): RillGrainOffset | undefined {
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
      grain: firstGrain.grain,
      offset,
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

  public toString() {
    return this.parts.map((part) => part.toString()).join("");
  }
}

export class RillGrainPointInTimePart {
  public constructor(
    public readonly prefix: string,
    public readonly grains: RillGrain[],
  ) {}

  public toString() {
    const grainLabels = this.grains
      .map((grain) => {
        const grainPrefix = grain.num ? grain.num : "";
        return `${grainPrefix}${grain.grain}`;
      })
      .join("");
    return `${this.prefix}${grainLabels}`;
  }
}

export class RillLabelledPointInTime implements RillPointInTimeVariant {
  public constructor(private readonly label: RillTimeLabel) {}

  public static postProcessor([label]: string[]) {
    return new RillLabelledPointInTime(label.toLowerCase() as RillTimeLabel);
  }

  public getGrain(): V1TimeGrain | undefined {
    return undefined;
  }

  public toString() {
    return this.label;
  }
}

export class RillAbsoluteTime implements RillPointInTimeVariant {
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

  public getGrain(): V1TimeGrain | undefined {
    return undefined;
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
type RillGrainOffset = {
  grain: string | undefined;
  offset: number;
};

function capitalizeFirstChar(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}
