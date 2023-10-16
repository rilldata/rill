export function getNextQuarterHour(): Date {
  const MS_PER_MINUTE = 60000;
  const MINUTES_PER_QUARTER_HOUR = 15;
  const MS_PER_QUARTER_HOUR = MS_PER_MINUTE * MINUTES_PER_QUARTER_HOUR;

  const currentTime = new Date();
  return new Date(
    Math.ceil(currentTime.getTime() / MS_PER_QUARTER_HOUR) * MS_PER_QUARTER_HOUR
  );
}

export function formatTime(date: Date): string {
  let hours = date.getHours();
  const minutes = date.getMinutes().toString().padStart(2, "0");
  const period = hours >= 12 ? "pm" : "am";

  hours = hours % 12;
  hours = hours ? hours : 12;

  return `${hours}:${minutes}${period}`;
}
