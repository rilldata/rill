import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config.ts";
import { isGrainBigger } from "@rilldata/web-common/lib/time/grains";
import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DateTime, Duration } from "luxon";
import type { DateObjectUnits } from "luxon/src/datetime";
import {
  getMinGrain,
  grainAliasToDateTimeUnit,
  GrainAliasToV1TimeGrain,
  V1TimeGrainToDateTimeUnit,
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

export type RillTimeAsOfLabel = {
  label: RillTimeLabel;
  snap: string | undefined;
  offset: number;
};

export class RillTime {
  public isComplete: boolean = false;
  public timezone: string | undefined;
  public anchorOverrides: RillPointInTime[] = [];

  public readonly rangeGrain: V1TimeGrain | undefined;
  public byGrain: V1TimeGrain | undefined;
  public readonly isShorthandSyntax: boolean;
  public asOfLabel: RillTimeAsOfLabel | undefined = undefined;

  public isOldFormat = false;

  public constructor(public readonly interval: RillTimeInterval) {
    this.updateIsComplete();

    this.isShorthandSyntax =
      interval instanceof RillShorthandInterval ||
      interval instanceof RillPeriodToGrainInterval;
    this.rangeGrain = this.interval.getGrain();
    this.isOldFormat =
      interval instanceof RillLegacyIsoInterval ||
      interval instanceof RillLegacyDaxInterval ||
      interval instanceof RillAllTimeInterval;
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
    this.asOfLabel = this.getAsOfLabel();
    this.updateIsComplete();
    return this;
  }

  public getLabel() {
    const offset = this.getAnchorOverridesOffset();
    const [label, supported] = this.interval.getLabel(offset);
    return supported ? capitalizeFirstChar(label) : this.toString();
  }

  public overrideRef(override: RillPointInTime) {
    if (this.isOldFormat) return;

    const pointUsingRefIndex = this.anchorOverrides.findIndex((pt) =>
      pt.hasLabelledPart(),
    );
    if (pointUsingRefIndex >= 0) {
      this.withAnchorOverrides([
        ...this.anchorOverrides.slice(0, pointUsingRefIndex),
        override,
        ...this.anchorOverrides.slice(pointUsingRefIndex + 1),
      ]);
    } else {
      this.withAnchorOverrides([...this.anchorOverrides, override]);
    }
  }

  public isAbsoluteTime() {
    return this.interval instanceof RillIsoInterval;
  }

  public toString() {
    let timeRange = this.interval.toString();

    timeRange += this.anchorOverrides
      .map((anchor) => ` as of ${anchor.toString()}`)
      .join("");

    if (this.byGrain) {
      timeRange += ` by ${this.byGrain}`;
    }

    if (this.timezone) {
      timeRange += ` tz ${this.timezone}`;
    }

    return timeRange;
  }

  private updateIsComplete() {
    const offset = this.getAnchorOverridesOffset();
    this.isComplete = this.interval.isComplete(offset, this.asOfLabel);
  }

  private getAnchorOverridesOffset(): Duration {
    let offset = Duration.fromObject({});

    this.anchorOverrides.forEach((anchor) => {
      offset = offset.plus(anchor.offset);
    });

    return offset;
  }

  private getAsOfLabel(): RillTimeAsOfLabel | undefined {
    const labelledAnchor = this.anchorOverrides.find((anchor) =>
      anchor.hasLabelledPart(),
    );
    if (!labelledAnchor) return undefined;

    const labelledPart = labelledAnchor.getLabelledPart();
    if (!labelledPart || labelledPart.snaps.length > 1) return undefined;
    const labelledPoint = labelledPart.point as RillLabelledPointInTime;

    const snap = labelledPart.snaps[0];
    const offsetForSnap =
      labelledAnchor.offset[
        V1TimeGrainToDateTimeUnit[GrainAliasToV1TimeGrain[snap]] + "s"
      ] ?? 0;

    return {
      label: labelledPoint.label,
      snap,
      offset: offsetForSnap,
    };
  }
}

interface RillTimeInterval {
  isComplete(
    offset: Duration,
    asOfLabel: RillTimeAsOfLabel | undefined,
  ): boolean;
  getLabel(offset: Duration): [label: string, supported: boolean];
  getGrain(): V1TimeGrain | undefined;
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

  public isComplete(
    offset: Duration,
    asOfLabel: RillTimeAsOfLabel | undefined,
  ) {
    return this.expandedInterval.isComplete(offset, asOfLabel);
  }

  public getLabel(offset: Duration): [label: string, supported: boolean] {
    return this.expandedInterval.getLabel(offset);
  }

  public getGrain() {
    return this.expandedInterval.getGrain();
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

  public isComplete(
    offset: Duration,
    asOfLabel: RillTimeAsOfLabel | undefined,
  ) {
    return this.expandedInterval.isComplete(offset, asOfLabel);
  }

  public getLabel(): [label: string, supported: boolean] {
    const grain = grainAliasToDateTimeUnit(this.grain as any);
    return [`${grain} to date`, true];
  }

  public getGrain() {
    return GrainAliasToV1TimeGrain[this.grain];
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

  public getGrain() {
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

  public isComplete(
    offset: Duration,
    asOfLabel: RillTimeAsOfLabel | undefined,
  ) {
    const endOffset = this.end.offset.plus(offset).toObject();
    const grains = Object.keys(endOffset);
    if (grains.length !== 1)
      return this.end.hasSnap() || !!asOfLabel?.snap || grains.length > 0;

    const grain = grains[0];
    const offsetAmount = endOffset[grain];
    if (offsetAmount < 0) return true;
    else if (offsetAmount > 0) return false;

    return this.end.hasSnap() || asOfLabel?.snap === grain;
  }

  public getLabel(offset: Duration): [label: string, supported: boolean] {
    let startOffset = this.start.offset.toObject();
    let endOffset = this.end.offset.toObject();
    const parentOffset = offset.toObject();

    if (this.start?.parts?.[0]?.point instanceof RillAbsoluteTime) {
      return ["Custom", true];
    }

    if (
      Object.keys(startOffset).length > 1 ||
      Object.keys(endOffset).length > 1 ||
      Object.keys(parentOffset).length > 1
    ) {
      return ["", false];
    }

    const startGrain = Object.keys(startOffset)[0];
    const endGrain = Object.keys(endOffset)[0];
    if (startGrain && endGrain && startGrain !== endGrain) {
      return ["", false];
    }

    const grain = startGrain || endGrain || "";

    const offsetGrain = Object.keys(parentOffset)[0];
    if (
      isGrainBigger(
        GrainAliasToV1TimeGrain[offsetGrain],
        GrainAliasToV1TimeGrain[grain],
      )
    ) {
      return ["", false];
    }
    startOffset = this.start.offset.plus(offset).toObject();
    endOffset = this.end.offset.plus(offset).toObject();

    const startOffsetAmount = startOffset[grain] ?? 0;
    const endOffsetAmount = endOffset[grain] ?? 0;
    const numDiff = Math.abs(startOffsetAmount - endOffsetAmount);

    const grainSingular = grain.replace(/s$/, "");
    const grainSuffix = numDiff > 1 ? "s" : "";
    const grainPrefix = numDiff ? numDiff + " " : "";
    const grainLabel = `${grainPrefix}${grainSingular}${grainSuffix}`;

    if (startOffsetAmount === 0 || startOffsetAmount === 1) {
      if (numDiff === 1) {
        const prefix = startOffsetAmount === 0 ? "this" : "next";
        return [`${prefix} ${grainSingular}`, true];
      }
      return [`next ${grainLabel}`, true];
    }

    if (endOffsetAmount === 0 || endOffsetAmount === 1) {
      if (numDiff === 1) {
        const prefix = endOffsetAmount === 1 ? "this" : "previous";
        return [`${prefix} ${grainSingular}`, true];
      }
      return [`last ${grainLabel}`, true];
    }

    return ["", false];
  }

  public getGrain() {
    const startRangeGrain = this.start.getGrain();
    const endRangeGrain =
      typeof this.end?.getGrain === "function"
        ? this.end.getGrain()
        : "TIME_GRAIN_DAY";
    const rangeGrain = getMinGrain(startRangeGrain, endRangeGrain);
    return rangeGrain;
  }

  public toString() {
    return `${this.start.toString()} to ${this.end.toString()}`;
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
    return ["Custom", true];
  }

  public getGrain() {
    return undefined;
  }

  public toString() {
    let timeRange = this.start.toString();
    if (this.end) {
      timeRange += ` to ${this.end.toString()}`;
    }
    return timeRange;
  }
}

export class RillAllTimeInterval implements RillTimeInterval {
  public isComplete() {
    return false;
  }

  public getLabel(): [label: string, supported: boolean] {
    return ["All time", true];
  }

  public getGrain() {
    return undefined;
  }

  public toString() {
    return "inf";
  }
}

export class RillLegacyIsoInterval implements RillTimeInterval {
  public constructor(
    private readonly dateGrains: RillGrain[],
    private readonly timeGrains: RillGrain[],
  ) {}

  public isComplete() {
    return false;
  }

  public getLabel(): [label: string, supported: boolean] {
    const isoDuration = this.toString();
    const label = "Last " + humaniseISODuration(isoDuration, false);
    return [label, true];
  }

  public getGrain() {
    return undefined;
  }

  public toString() {
    const dateParts = this.dateGrains.map(
      ({ grain, num }) => `${num!}${grain.toUpperCase()}`,
    );
    const datePart = "P" + dateParts.join("");
    const timeParts = this.timeGrains.map(
      ({ grain, num }) => `${num!}${grain.toUpperCase()}`,
    );
    const timePart = timeParts.length > 0 ? "T" + timeParts.join("") : "";
    return datePart + timePart;
  }
}

export class RillLegacyDaxInterval implements RillTimeInterval {
  public constructor(private readonly name: string) {}

  public isComplete() {
    return false;
  }

  public getLabel(): [label: string, supported: boolean] {
    const entry = DEFAULT_TIME_RANGES[this.name];
    if (!entry) return ["", false];
    return [entry.label, true];
  }

  public getGrain() {
    return undefined;
  }

  public toString() {
    return this.name;
  }
}

export class RillPointInTime {
  public readonly offset: Duration;

  public constructor(public readonly parts: RillPointInTimeWithSnap[]) {
    let offset = Duration.fromObject({});
    parts.forEach((part) => {
      offset = offset.plus(part.offset);
    });
    this.offset = offset.normalize();
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

  public getLabelledPart() {
    return this.parts.find((p) => p.point instanceof RillLabelledPointInTime);
  }

  public hasSnap() {
    return this.parts.some((p) => p.snaps.length);
  }

  public toString() {
    return this.parts.map((part) => part.toString()).join("");
  }
}

export class RillPointInTimeWithSnap {
  public readonly offset = Duration.fromObject({});

  public constructor(
    public readonly point: RillPointInTimeVariant,
    public snaps: string[],
  ) {
    if (this.point instanceof RillGrainPointInTime) {
      this.offset = this.point.offset;
    }
  }

  public toString() {
    return `${this.point.toString()}${this.snaps.map((s) => "/" + s).join("")}`;
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
  public readonly offset: Duration;

  public constructor(public readonly parts: RillGrainPointInTimePart[]) {
    let offset = Duration.fromObject({});
    parts.forEach((part) => {
      if (part.prefix === "+") {
        offset = offset.plus(part.offset);
      } else {
        offset = offset.minus(part.offset);
      }
    });
    this.offset = offset.normalize();
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
  public readonly offset: Duration;

  public constructor(
    public readonly prefix: string,
    public readonly grains: RillGrain[],
  ) {
    let offset = Duration.fromObject({});
    grains.forEach(({ grain, num }) => {
      const luxonGrain =
        V1TimeGrainToDateTimeUnit[GrainAliasToV1TimeGrain[grain]];
      if (!luxonGrain || !num) return;
      offset = offset.plus({ [luxonGrain]: num });
    });
    this.offset = offset.normalize();
  }

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
  public constructor(public readonly label: RillTimeLabel) {}

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
    return new RillAbsoluteTime(args.flat().join(""));
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

export function capitalizeFirstChar(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}
