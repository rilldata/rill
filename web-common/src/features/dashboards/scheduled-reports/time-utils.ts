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
