export class RillTime {
  public readonly isComplete: boolean;
  public readonly end: RillTimeModifier;

  public constructor(
    public readonly start: RillTimeModifier,
    end: RillTimeModifier,
    public readonly modifier: RillTimeRangeModifier | undefined,
  ) {
    this.end = end ?? RillTimeModifier.now();
    this.isComplete =
      this.end.type === RillTimeModifierType.Custom ||
      this.end.truncate !== undefined;
  }

  public getLabel() {
    const start = capitalizeFirstChar(this.start.getLabel(this.isComplete));
    const completeSuffix = this.isComplete ? "complete" : "incomplete";
    return `${start}, ${completeSuffix}`;
  }
}

export enum RillTimeModifierType {
  Now = "Now",
  Earliest = "Earliest",
  Latest = "Latest",
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
export class RillTimeModifier {
  public truncate: RillTimeGrain | undefined = undefined;

  public constructor(
    public readonly type: RillTimeModifierType,
    public readonly grain: RillTimeGrain | undefined = undefined,
  ) {}

  public static now() {
    return new RillTimeModifier(RillTimeModifierType.Now);
  }
  public static earliest() {
    return new RillTimeModifier(RillTimeModifierType.Earliest);
  }
  public static latest() {
    return new RillTimeModifier(RillTimeModifierType.Latest);
  }
  public static custom(grain: RillTimeGrain) {
    return new RillTimeModifier(RillTimeModifierType.Custom, grain);
  }

  public withTruncate(truncate: RillTimeGrain) {
    this.truncate = truncate;
    return this;
  }

  public getLabel(isComplete: boolean) {
    const grain = this.grain ?? this.truncate;
    if (!grain) {
      return RillTimeModifierType.Earliest.toString();
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
    const completenessOffset = isComplete ? 0 : 1;
    return `last ${-grain.count + completenessOffset} ${unit}s`;
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
  timeRangeGrain: RillTimeRangeGrain | undefined;
  timeZone: string | undefined;
  at: RillTimeModifier | undefined;
};

function capitalizeFirstChar(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}
