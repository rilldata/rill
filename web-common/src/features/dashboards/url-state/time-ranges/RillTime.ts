export enum RillTimeType {
  Unknown = "Unknown",
  Latest = "Latest",
  PreviousPeriod = "Previous period",
  PeriodToDate = "Period To Date",
}

export class RillTime {
  public timeRange: string;
  public readonly isComplete: boolean;
  public readonly end: RillTimeAnchor;
  public readonly type: RillTimeType;

  public constructor(
    public readonly start: RillTimeAnchor,
    end: RillTimeAnchor,
    public readonly timeRangeGrain: RillTimeRangeGrain | undefined,
    public readonly modifier: RillTimeRangeModifier | undefined,
  ) {
    this.type = start.getType();

    this.end = end ?? RillTimeAnchor.now();
    this.isComplete =
      this.end.type === RillTimeAnchorType.Relative ||
      this.end.truncate !== undefined;
  }

  public getLabel() {
    if (this.type === RillTimeType.Unknown || !!this.modifier) {
      return this.timeRange;
    }

    const start = capitalizeFirstChar(this.start.getLabel());
    const hasNonStandardStart =
      this.start.type === RillTimeAnchorType.Custom || !!this.start.offset;
    const hasNonStandardEnd =
      this.end &&
      ((this.end.type === RillTimeAnchorType.Relative &&
        this.end.grain &&
        this.end.grain.count !== 0) ||
        this.end.type === RillTimeAnchorType.Custom ||
        !!this.end.offset);
    if (hasNonStandardStart || hasNonStandardEnd) {
      return this.timeRange;
    }

    if (this.isComplete) return start;
    return `${start}, incomplete`;
  }
}

export enum RillTimeAnchorType {
  Now = "Now",
  Earliest = "Earliest",
  Latest = "Latest",
  Relative = "Relative",
  Custom = "Custom",
}

const GrainToUnit = {
  s: "second",
  m: "minute",
  h: "hour",
  d: "day",
  D: "day",
  W: "week",
  M: "month",
  Q: "Quarter",
  Y: "year",
};
export const InvalidTime = "Invalid";
export class RillTimeAnchor {
  public truncate: RillTimeGrain | undefined = undefined;
  public absolute: string | undefined = undefined;
  public grain: RillTimeGrain | undefined = undefined;
  public offset: RillTimeGrain | undefined = undefined;

  public constructor(public readonly type: RillTimeAnchorType) {}

  public static now() {
    return new RillTimeAnchor(RillTimeAnchorType.Now);
  }
  public static earliest() {
    return new RillTimeAnchor(RillTimeAnchorType.Earliest);
  }
  public static latest() {
    return new RillTimeAnchor(RillTimeAnchorType.Latest);
  }
  public static relative(grain: RillTimeGrain) {
    return new RillTimeAnchor(RillTimeAnchorType.Relative).withGrain(grain);
  }
  public static absolute(time: string) {
    return new RillTimeAnchor(RillTimeAnchorType.Custom).withAbsolute(time);
  }

  public withGrain(grain: RillTimeGrain) {
    this.grain = grain;
    return this;
  }

  public withOffset(grain: RillTimeGrain) {
    this.offset = grain;
    return this;
  }

  public withAbsolute(time: string) {
    this.absolute = time;
    return this;
  }

  public withTruncate(truncate: RillTimeGrain) {
    this.truncate = truncate;
    return this;
  }

  public getLabel() {
    const grain = this.grain ?? this.truncate;
    if (!grain) {
      return RillTimeAnchorType.Earliest.toString();
    }

    const unit = GrainToUnit[grain.grain];
    if (!unit) return InvalidTime;

    if (grain.count === 0) {
      if (unit === "day") return "today";
      return `${unit} to date`;
    }

    if (grain.count > 0) return InvalidTime;

    if (grain.count === -1) {
      return `previous ${unit}`;
    }
    return `last ${-grain.count} ${unit}s`;
  }

  public getType() {
    const grain = this.grain ?? this.truncate;
    if (!grain || grain.count > 0) {
      return RillTimeType.Unknown;
    }

    if (grain.count === 0) {
      return RillTimeType.PeriodToDate;
    }
    if (grain.count === -1) {
      return RillTimeType.PreviousPeriod;
    }
    return RillTimeType.Latest;
  }
}

export type RillTimeGrain = {
  grain: string;
  count: number;
};
export type RillTimeRangeGrain = {
  grain: string;
  isComplete: boolean;
};

export type RillTimeRangeModifier = {
  timeZone: string | undefined;
  at: RillTimeAnchor | undefined;
};

function capitalizeFirstChar(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}
