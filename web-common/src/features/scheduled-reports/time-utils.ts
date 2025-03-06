import { DateTime } from "luxon";
import {
  getAbbreviationForIANA,
  getLocalIANA,
  getUTCIANA,
} from "../../lib/time/timezone";

export enum ReportFrequency {
  Daily = "Daily",
  Weekdays = "Weekdays",
  Weekly = "Weekly",
  Monthly = "Monthly",
  Custom = "Custom",
}

export function getTodaysDayOfWeek(): string {
  return DateTime.now().toLocaleString({ weekday: "long" });
}

export function getTodaysDayOfMonth(): number {
  return DateTime.now().day;
}

export function getNextQuarterHour(): DateTime {
  const now = DateTime.local();
  const nextQuarter = now.plus({ minutes: 15 - (now.minute % 15) });
  return nextQuarter.startOf("minute");
}

export function getTimeIn24FormatFromDateTime(dateTime: DateTime): string {
  return dateTime.toFormat("HH:mm");
}

const weekDayMap: Record<string, number> = {
  Sunday: 0,
  Monday: 1,
  Tuesday: 2,
  Wednesday: 3,
  Thursday: 4,
  Friday: 5,
  Saturday: 6,
};
export function convertFormValuesToCronExpression(
  frequency: ReportFrequency,
  dayOfWeek: string,
  timeOfDay: string,
  dayOfMonth: number,
): string {
  const [hour, minute] = timeOfDay.split(":").map(Number);
  let cronExpr = `${minute} ${hour} `;

  switch (frequency) {
    case ReportFrequency.Daily:
      cronExpr += "* * *";
      break;
    case ReportFrequency.Weekdays:
      cronExpr += "* * 1-5";
      break;
    case ReportFrequency.Weekly: {
      cronExpr += `* * ${weekDayMap[dayOfWeek]}`;
      break;
    }
    case ReportFrequency.Monthly:
      cronExpr += `${dayOfMonth} * *`;
      break;
  }

  return cronExpr;
}

export function getFrequencyFromCronExpression(
  cronExpr: string,
): ReportFrequency {
  const [, , dayOfMonth, month, dayOfWeek] = cronExpr.split(" ");

  if (dayOfMonth === "*" && month === "*") {
    if (dayOfWeek === "*") {
      return ReportFrequency.Daily;
    } else if (dayOfWeek === "1-5") {
      return ReportFrequency.Weekdays;
    } else {
      return ReportFrequency.Weekly;
    }
  }
  if (month === "*" && dayOfWeek === "*") {
    return ReportFrequency.Monthly;
  }

  return ReportFrequency.Custom;
}

export function getDayOfWeekFromCronExpression(cronExpr: string): string {
  const [, , , , dayOfWeek] = cronExpr.split(" ");

  switch (dayOfWeek) {
    case "0":
      return "Sunday";
    case "1":
      return "Monday";
    case "2":
      return "Tuesday";
    case "3":
      return "Wednesday";
    case "4":
      return "Thursday";
    case "5":
      return "Friday";
    case "6":
      return "Saturday";
    default:
      return "";
  }
}

export function getTimeOfDayFromCronExpression(cronExpr: string): string {
  const [minute, hour, , ,] = cronExpr.split(" ");
  return `${hour}:${minute === "0" ? "00" : minute}`;
}

export function getDayOfMonthFromCronExpression(cronExpr: string): number {
  const [, , dayOfMonth] = cronExpr.split(" ");
  const dayOfMonthAsNum = Number(dayOfMonth);
  if (Number.isNaN(dayOfMonthAsNum)) return 1;
  return dayOfMonthAsNum;
}

export function makeTimeZoneOptions(availableTimeZones: string[] | undefined) {
  const userLocalIANA = getLocalIANA();
  const UTCIana = getUTCIANA();
  const currentDate = new Date();

  if (!availableTimeZones) {
    return [
      {
        value: userLocalIANA,
        label: getAbbreviationForIANA(currentDate, userLocalIANA) + " (Local)",
      },
      {
        value: UTCIana,
        label: getAbbreviationForIANA(currentDate, UTCIana),
      },
    ];
  }

  // Add local time and UTC to available time zones
  const extendedTimeZones = [userLocalIANA, UTCIana, ...availableTimeZones];

  // Create a map to deduplicate time zones
  const deduplicatedTimeZones = new Map();

  extendedTimeZones.forEach((z) => {
    const abbreviation = getAbbreviationForIANA(currentDate, z);
    if (!deduplicatedTimeZones.has(abbreviation)) {
      deduplicatedTimeZones.set(abbreviation, z);
    }
  });

  // Convert the map back to an array of options
  return Array.from(deduplicatedTimeZones).map(([abbreviation, value]) => {
    return {
      value: value,
      label: abbreviation + (value === userLocalIANA ? " (Local)" : ""),
    };
  });
}
