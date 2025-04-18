export enum RillTimeType {
  Unknown = "Unknown",
  Latest = "Latest",
  PreviousPeriod = "Previous period",
  PeriodToDate = "Period To Date",
}

const absTimeRegex =
  /(?<year>\d{4})(-(?<month>\d{2})(-(?<day>\d{2})(T(?<hour>\d{2})(:(?<minute>\d{2})(:(?<second>\d{2})Z)?)?)?)?)?/;

export class RillTime {
  public timeRange: string;
  public readonly isComplete: boolean;
  public readonly type = RillTimeType.Unknown; // TODO

  public constructor(
    public readonly start: RillTimePart[],
    public readonly end: RillTimePart[] | undefined,
    public readonly timeRangeGrain: string | undefined,
    public readonly timezone: string | undefined,
  ) {
    this.isComplete = end?.[0]?.isComplete ?? start[0]?.isComplete ?? false;
  }

  public getLabel() {
    if (this.end) return this.timeRange; // TODO: what would the labels be here?

    let range = this.start.map((p) => p.getLabel()).join(" of ");

    if (this.timeRangeGrain) {
      range += ` by ${this.timeRangeGrain}`;
    }

    if (this.timezone) {
      range += ` @{${this.timezone}}`;
    }

    return range;
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
      range += ` @{${this.timezone}}`;
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
  private readonly time: Date;
  public isComplete = true; // TODO: can this be anything else?

  public constructor(timeStr: string) {
    const absTimeMatch = absTimeRegex.exec(timeStr);
    if (!absTimeMatch) {
      this.time = new Date(0);
      return;
    }

    const year = Number(absTimeMatch.groups?.year ?? "0");
    const month = Number(absTimeMatch.groups?.month ?? "1");
    const day = Number(absTimeMatch.groups?.day ?? "1");
    const hour = Number(absTimeMatch.groups?.hour ?? "0");
    const minute = Number(absTimeMatch.groups?.minute ?? "0");
    const second = Number(absTimeMatch.groups?.second ?? "0");

    this.time = new Date(year, month, day, hour, minute, second, 0);
  }

  public getLabel() {
    return this.time.toLocaleTimeString();
  }

  public toString() {
    return this.time.toISOString();
  }
}

export class RillTimeLabelledAnchor implements RillTimePart {
  public isComplete = false; // TODO: can this be anything else?

  public constructor(public readonly label: string) {}

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
    const grainPart = capitalizeFirstChar(GrainToUnit[this.grain]);
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
        return `Previous ${grainLabel}`;

      case "-":
        if (this.num === 1) {
          return `Previous ${grainPart}`;
        }
        return `${grainLabel} in the past`;

      case "+":
        if (this.num === 1) {
          return `Next ${grainPart}`;
        }
        return `${grainLabel} in the future`;

      case "<":
        return `First ${grainLabel} in the future`;

      case ">":
        return `Last ${grainLabel} in the future`;
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
    // TODO
    return `${this.from} to ${this.to}`;
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
