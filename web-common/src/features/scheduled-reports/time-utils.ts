import { DateTime } from "luxon";
import {
  getAbbreviationForIANA,
  getLocalIANA,
  getUTCIANA,
} from "../../lib/time/timezone";

export function getTodaysDayOfWeek(): string {
  return DateTime.now().toLocaleString({ weekday: "long" });
}

export function getNextQuarterHour(): DateTime {
  const now = DateTime.local();
  const nextQuarter = now.plus({ minutes: 15 - (now.minute % 15) });
  return nextQuarter.startOf("minute");
}

export function getTimeIn24FormatFromDateTime(dateTime: DateTime): string {
  return dateTime.toFormat("HH:mm");
}

export function convertFormValuesToCronExpression(
  frequency: string,
  dayOfWeek: string,
  timeOfDay: string,
): string {
  const [hour, minute] = timeOfDay.split(":").map(Number);
  let cronExpr = `${minute} ${hour} `;

  switch (frequency) {
    case "Daily":
      cronExpr += "* * *";
      break;
    case "Weekdays":
      cronExpr += "* * 1-5";
      break;
    case "Weekly": {
      const weekDayMap: Record<string, number> = {
        Sunday: 0,
        Monday: 1,
        Tuesday: 2,
        Wednesday: 3,
        Thursday: 4,
        Friday: 5,
        Saturday: 6,
      };
      cronExpr += `* * ${weekDayMap[dayOfWeek]}`;
      break;
    }
  }

  return cronExpr;
}

export function getFrequencyFromCronExpression(cronExpr: string): string {
  const [, , dayOfMonth, month, dayOfWeek] = cronExpr.split(" ");

  if (dayOfMonth === "*" && month === "*") {
    if (dayOfWeek === "*") {
      return "Daily";
    } else if (dayOfWeek === "1-5") {
      return "Weekdays";
    } else {
      return "Weekly";
    }
  }

  return "Custom";
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
  return `${hour}:${minute}`;
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
