import { DateTime } from "luxon";

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

export function convertToCron(
  frequency: string,
  dayOfWeek: string,
  timeOfDay: string
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

export function getFrequencyFromCron(cronExpr: string): string {
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

export function getDayOfWeekFromCron(cronExpr: string): string {
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

export function getTimeOfDayFromCron(cronExpr: string): string {
  const [minute, hour, , ,] = cronExpr.split(" ");
  return `${hour}:${minute}`;
}
