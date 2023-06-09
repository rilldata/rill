export class DateTime {
  public static parseDateTime(
    date: Date | DateTime | string | number,
    format: string = "YYYY-MM-DD",
    lang: string = "en-US"
  ): Date {
    if (!date) return new Date(NaN);

    if (date instanceof Date) return new Date(date);
    if (date instanceof DateTime) return date.clone().toJSDate();

    if (/^-?\d{10,}$/.test(date as string))
      return DateTime.getDateZeroTime(new Date(Number(date)));

    if (typeof date === "string") {
      const matches = [];
      let m = null;

      // tslint:disable-next-line: no-conditional-assignment
      while ((m = DateTime.regex.exec(format)) != null) {
        if (m[1] === "\\") continue; // delete when regexp lookbehind

        matches.push(m);
      }

      if (matches.length) {
        const datePattern = {
          year: null,
          month: null,
          shortMonth: null,
          longMonth: null,
          day: null,
          value: "",
        };

        if (matches[0].index > 0) {
          datePattern.value += ".*?";
        }

        for (const [k, match] of Object.entries(matches)) {
          const key = Number(k);

          const { group, pattern } = DateTime.formatPatterns(match[0], lang);

          datePattern[group] = key + 1;
          datePattern.value += pattern;

          datePattern.value += ".*?"; // any delimiters
        }

        const dateRegex = new RegExp(`^${datePattern.value}$`);

        if (dateRegex.test(date)) {
          const d = dateRegex.exec(date);

          const year = Number(d[datePattern.year]);
          let month = null;

          if (datePattern.month) {
            month = Number(d[datePattern.month]) - 1;
          } else if (datePattern.shortMonth) {
            month = DateTime.shortMonths(lang).indexOf(
              d[datePattern.shortMonth]
            );
          } else if (datePattern.longMonth) {
            month = DateTime.longMonths(lang).indexOf(d[datePattern.longMonth]);
          }

          const day = Number(d[datePattern.day]) || 1;

          return new Date(year, month, day, 0, 0, 0, 0);
        }
      }
    }

    return DateTime.getDateZeroTime(new Date(date));
  }

  public static convertArray(
    array: Array<Date | Date[] | string | string[]>,
    format: string
  ): Array<DateTime | DateTime[]> {
    return array.map((d) => {
      if (d instanceof Array) {
        return (d as Array<Date | string>).map(
          (d1) => new DateTime(d1, format)
        );
      }
      return new DateTime(d, format);
    });
  }

  public static getDateZeroTime(date: Date): Date {
    return new Date(
      date.getFullYear(),
      date.getMonth(),
      date.getDate(),
      0,
      0,
      0,
      0
    );
  }

  // replace to regexp lookbehind when most popular browsers will support
  // https://caniuse.com/#feat=js-regexp-lookbehind
  // /(?<!\\)(Y{2,4}|M{1,4}|D{1,2}|d{1,4}])/g
  private static regex: RegExp = /(\\)?(Y{2,4}|M{1,4}|D{1,2}|d{1,4})/g;

  private static readonly MONTH_JS = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11];

  private static shortMonths(lang): string[] {
    return DateTime.MONTH_JS.map((x) =>
      new Date(2019, x).toLocaleString(lang, { month: "short" })
    );
  }

  private static longMonths(lang): string[] {
    return DateTime.MONTH_JS.map((x) =>
      new Date(2019, x).toLocaleString(lang, { month: "long" })
    );
  }

  private static formatPatterns(token, lang) {
    switch (token) {
      case "YY":
      case "YYYY":
        return {
          group: "year",
          pattern: `(\\d{${token.length}})`,
        };

      case "M":
        return {
          group: "month",
          pattern: "(\\d{1,2})",
        };

      case "MM":
        return {
          group: "month",
          pattern: "(\\d{2})",
        };

      case "MMM":
        return {
          group: "shortMonth",
          pattern: `(${DateTime.shortMonths(lang).join("|")})`,
        };

      case "MMMM":
        return {
          group: "longMonth",
          pattern: `(${DateTime.longMonths(lang).join("|")})`,
        };

      case "D":
        return {
          group: "day",
          pattern: "(\\d{1,2})",
        };

      case "DD":
        return {
          group: "day",
          pattern: "(\\d{2})",
        };
    }
  }

  protected lang: string;

  private dateInstance: Date;

  constructor(
    date: Date | DateTime | number | string = null,
    format: object | string = null,
    lang: string = "en-US"
  ) {
    if (typeof format === "object" && format !== null) {
      // tslint:disable-next-line: max-line-length
      this.dateInstance = (format as any).parse(
        date instanceof DateTime ? date.clone().toJSDate() : date
      );
    } else if (typeof format === "string") {
      this.dateInstance = DateTime.parseDateTime(date, format, lang);
    } else if (date) {
      this.dateInstance = DateTime.parseDateTime(date);
    } else {
      this.dateInstance = DateTime.parseDateTime(new Date());
    }

    this.lang = lang;
  }

  public toJSDate(): Date {
    return this.dateInstance;
  }

  public toLocaleString(
    arg0: string,
    arg1: Intl.DateTimeFormatOptions
  ): string {
    return this.dateInstance.toLocaleString(arg0, arg1);
  }

  public toDateString(): string {
    return this.dateInstance.toDateString();
  }

  public getSeconds(): number {
    return this.dateInstance.getSeconds();
  }

  public getDay(): number {
    return this.dateInstance.getDay();
  }

  public getTime(): number {
    return this.dateInstance.getTime();
  }

  public getDate(): number {
    return this.dateInstance.getDate();
  }

  public getMonth(): number {
    return this.dateInstance.getMonth();
  }

  public getFullYear(): number {
    return this.dateInstance.getFullYear();
  }

  public setMonth(arg: number): number {
    return this.dateInstance.setMonth(arg);
  }

  public setHours(
    hours: number = 0,
    minutes: number = 0,
    seconds: number = 0,
    ms: number = 0
  ) {
    this.dateInstance.setHours(hours, minutes, seconds, ms);
  }

  public setSeconds(arg: number): number {
    return this.dateInstance.setSeconds(arg);
  }

  public setDate(arg: number): number {
    return this.dateInstance.setDate(arg);
  }

  public setFullYear(arg: number): number {
    return this.dateInstance.setFullYear(arg);
  }

  public getWeek(firstDay: number): number {
    const target = new Date(this.timestamp());
    const dayNr = (this.getDay() + (7 - firstDay)) % 7;
    target.setDate(target.getDate() - dayNr);
    const startWeekday = target.getTime();
    target.setMonth(0, 1);
    if (target.getDay() !== firstDay) {
      target.setMonth(0, 1 + ((4 - target.getDay() + 7) % 7));
    }
    return 1 + Math.ceil((startWeekday - target.getTime()) / 604800000);
  }

  public clone(): DateTime {
    return new DateTime(this.toJSDate());
  }

  public isBetween(
    date1: DateTime,
    date2: DateTime,
    inclusivity = "()"
  ): boolean {
    switch (inclusivity) {
      default:
      case "()":
        return (
          this.timestamp() > date1.getTime() &&
          this.timestamp() < date2.getTime()
        );

      case "[)":
        return (
          this.timestamp() >= date1.getTime() &&
          this.timestamp() < date2.getTime()
        );

      case "(]":
        return (
          this.timestamp() > date1.getTime() &&
          this.timestamp() <= date2.getTime()
        );

      case "[]":
        return (
          this.timestamp() >= date1.getTime() &&
          this.timestamp() <= date2.getTime()
        );
    }
  }

  public isBefore(date: DateTime, unit = "seconds"): boolean {
    switch (unit) {
      case "second":
      case "seconds":
        return date.getTime() > this.getTime();

      case "day":
      case "days":
        return (
          new Date(
            date.getFullYear(),
            date.getMonth(),
            date.getDate()
          ).getTime() >
          new Date(
            this.getFullYear(),
            this.getMonth(),
            this.getDate()
          ).getTime()
        );

      case "month":
      case "months":
        return (
          new Date(date.getFullYear(), date.getMonth(), 1).getTime() >
          new Date(this.getFullYear(), this.getMonth(), 1).getTime()
        );

      case "year":
      case "years":
        return date.getFullYear() > this.getFullYear();
    }

    throw new Error("isBefore: Invalid unit!");
  }

  public isSameOrBefore(date: DateTime, unit = "seconds"): boolean {
    switch (unit) {
      case "second":
      case "seconds":
        return date.getTime() >= this.getTime();

      case "day":
      case "days":
        return (
          new Date(
            date.getFullYear(),
            date.getMonth(),
            date.getDate()
          ).getTime() >=
          new Date(
            this.getFullYear(),
            this.getMonth(),
            this.getDate()
          ).getTime()
        );

      case "month":
      case "months":
        return (
          new Date(date.getFullYear(), date.getMonth(), 1).getTime() >=
          new Date(this.getFullYear(), this.getMonth(), 1).getTime()
        );
    }

    throw new Error("isSameOrBefore: Invalid unit!");
  }

  public isAfter(date: DateTime, unit = "seconds"): boolean {
    switch (unit) {
      case "second":
      case "seconds":
        return this.getTime() > date.getTime();

      case "day":
      case "days":
        return (
          new Date(
            this.getFullYear(),
            this.getMonth(),
            this.getDate()
          ).getTime() >
          new Date(
            date.getFullYear(),
            date.getMonth(),
            date.getDate()
          ).getTime()
        );

      case "month":
      case "months":
        return (
          new Date(this.getFullYear(), this.getMonth(), 1).getTime() >
          new Date(date.getFullYear(), date.getMonth(), 1).getTime()
        );

      case "year":
      case "years":
        return this.getFullYear() > date.getFullYear();
    }

    throw new Error("isAfter: Invalid unit!");
  }

  public isSameOrAfter(date: DateTime, unit = "seconds"): boolean {
    switch (unit) {
      case "second":
      case "seconds":
        return this.getTime() >= date.getTime();

      case "day":
      case "days":
        return (
          new Date(
            this.getFullYear(),
            this.getMonth(),
            this.getDate()
          ).getTime() >=
          new Date(
            date.getFullYear(),
            date.getMonth(),
            date.getDate()
          ).getTime()
        );

      case "month":
      case "months":
        return (
          new Date(this.getFullYear(), this.getMonth(), 1).getTime() >=
          new Date(date.getFullYear(), date.getMonth(), 1).getTime()
        );
    }

    throw new Error("isSameOrAfter: Invalid unit!");
  }

  public isSame(date: DateTime, unit = "seconds"): boolean {
    switch (unit) {
      case "second":
      case "seconds":
        return this.getTime() === date.getTime();

      case "day":
      case "days":
        return (
          new Date(
            this.getFullYear(),
            this.getMonth(),
            this.getDate()
          ).getTime() ===
          new Date(
            date.getFullYear(),
            date.getMonth(),
            date.getDate()
          ).getTime()
        );

      case "month":
      case "months":
        return (
          new Date(this.getFullYear(), this.getMonth(), 1).getTime() ===
          new Date(date.getFullYear(), date.getMonth(), 1).getTime()
        );
    }

    throw new Error("isSame: Invalid unit!");
  }

  public add(duration: number, unit = "seconds"): DateTime {
    switch (unit) {
      case "second":
      case "seconds":
        this.setSeconds(this.getSeconds() + duration);
        break;

      case "day":
      case "days":
        this.setDate(this.getDate() + duration);
        break;

      case "month":
      case "months":
        this.setMonth(this.getMonth() + duration);
        break;
    }

    return this;
  }

  public subtract(duration: number, unit = "seconds"): DateTime {
    switch (unit) {
      case "second":
      case "seconds":
        this.setSeconds(this.getSeconds() - duration);
        break;

      case "day":
      case "days":
        this.setDate(this.getDate() - duration);
        break;

      case "month":
      case "months":
        this.setMonth(this.getMonth() - duration);
        break;
    }

    return this;
  }

  public diff(date: DateTime, unit = "seconds"): number {
    const oneDay = 1000 * 60 * 60 * 24;

    switch (unit) {
      default:
      case "second":
      case "seconds":
        return this.getTime() - date.getTime();

      case "day":
      case "days":
        return Math.round((this.timestamp() - date.getTime()) / oneDay);

      case "month":
      case "months":
      // @TODO
    }
  }

  public format(format: object | string, lang: string = "en-US"): string {
    if (typeof format === "object") {
      return (format as any).output(this.clone().toJSDate());
    }

    let response = "";

    const matches = [];
    let m = null;

    // tslint:disable-next-line: no-conditional-assignment
    while ((m = DateTime.regex.exec(format)) != null) {
      if (m[1] === "\\") continue; // delete when regexp lookbehind

      matches.push(m);
    }

    if (matches.length) {
      // add start line of tokens are not at the beginning
      if (matches[0].index > 0) {
        response += format.substring(0, matches[0].index);
      }

      for (const [k, match] of Object.entries(matches)) {
        const key = Number(k);
        response += this.formatTokens(match[0], lang);

        if (matches[key + 1]) {
          response += format.substring(
            match.index + match[0].length,
            matches[key + 1].index
          );
        }

        // add end line if tokens are not at the ending
        if (key === matches.length - 1) {
          response += format.substring(match.index + match[0].length);
        }
      }
    }

    // remove escape characters
    return response.replace(/\\/g, "");
  }

  private timestamp(): number {
    return new Date(
      this.getFullYear(),
      this.getMonth(),
      this.getDate(),
      0,
      0,
      0,
      0
    ).getTime();
  }

  private formatTokens(token, lang) {
    switch (token) {
      case "YY":
        return String(this.getFullYear()).slice(-2);
      case "YYYY":
        return String(this.getFullYear());

      case "M":
        return String(this.getMonth() + 1);
      case "MM":
        return `0${this.getMonth() + 1}`.slice(-2);
      case "MMM":
        return DateTime.shortMonths(lang)[this.getMonth()];
      case "MMMM":
        return DateTime.longMonths(lang)[this.getMonth()];

      case "D":
        return String(this.getDate());
      case "DD":
        return `0${this.getDate()}`.slice(-2);

      default:
        return "";
    }
  }
}
