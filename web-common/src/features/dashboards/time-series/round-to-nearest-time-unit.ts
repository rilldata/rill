import { DateTime, DateTimeUnit } from "luxon";
export function roundToNearestTimeUnit(date, unit: keyof DateTime) {
  const dateTime = DateTime.fromJSDate(date);
  if (!DateTime.isDateTime(dateTime)) {
    throw new Error("Invalid Luxon DateTime object");
  }

  const unitMap = {
    year: "month",
    month: "day",
    week: "day",
    day: "hour",
    hour: "minute",
    minute: "second",
    second: "milli",
  };
  // get smallest unit
  const smallerUnit = unitMap[unit];

  const smallestValue = dateTime.get(smallerUnit);
  let roundUp = false;
  if (smallerUnit === "milli") {
    roundUp = smallestValue >= 500;
  } else if (smallerUnit === "second") {
    roundUp = smallestValue >= 30;
  } else if (smallerUnit === "minute") {
    roundUp = smallestValue >= 30;
  } else if (smallerUnit === "hour") {
    roundUp = smallestValue >= 12;
  } else if (smallerUnit === "day") {
    roundUp = unit === "month" ? smallestValue >= 15 : smallestValue >= 3;
  } else if (smallerUnit === "week") {
    roundUp = smallestValue >= 3;
  } else if (smallerUnit === "month") {
    roundUp = smallestValue >= 6;
  }

  const unitValue = dateTime.get(unit);
  const roundedValue = roundUp ? unitValue + 1 : unitValue;
  const roundedDateTime = dateTime
    .startOf(unit as DateTimeUnit)
    .set({ [unit]: roundedValue });

  return roundedDateTime.toJSDate();
}
