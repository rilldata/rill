import { DateTime } from "luxon";

export function getTodaysDayOfWeek(): string {
  return DateTime.now().toLocaleString({ weekday: "long" });
}

export function getNextQuarterHour(): Date {
  const MS_PER_MINUTE = 60000;
  const MINUTES_PER_QUARTER_HOUR = 15;
  const MS_PER_QUARTER_HOUR = MS_PER_MINUTE * MINUTES_PER_QUARTER_HOUR;

  const currentTime = new Date();
  return new Date(
    Math.ceil(currentTime.getTime() / MS_PER_QUARTER_HOUR) * MS_PER_QUARTER_HOUR
  );
}

export function getTimeIn24FormatFromDate(date: Date): string {
  return `${date.getHours()}:${date.getMinutes().toString().padStart(2, "0")}`;
}

export function formatTime(date: Date): string {
  let hours = date.getHours();
  const minutes = date.getMinutes().toString().padStart(2, "0");
  const period = hours >= 12 ? "pm" : "am";

  hours = hours % 12;
  hours = hours ? hours : 12;

  return `${hours}:${minutes}${period}`;
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
