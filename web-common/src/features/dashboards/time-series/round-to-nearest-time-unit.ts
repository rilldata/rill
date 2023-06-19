import { DateTime, DateTimeUnit } from "luxon";
export function roundToNearestTimeUnit(
  date,
  unit: DateTimeUnit | keyof DateTime
) {
  const dateTime = DateTime.fromJSDate(date);
  if (!DateTime.isDateTime(dateTime)) {
    throw new Error("Invalid Luxon DateTime object");
  }

  const unitMap = {
    year: "month",
    month: "day",
    week: "weekday",
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
    roundUp = smallestValue >= 15;
  } else if (smallerUnit === "week") {
    roundUp = smallestValue >= 3;
  } else if (smallerUnit === "weekday") {
    roundUp = smallestValue >= 3;
  } else if (smallerUnit === "month") {
    roundUp = smallestValue >= 6;
  }

  const unitValue = dateTime.get(unit as keyof DateTime);
  const roundedValue = roundUp ? unitValue + 1 : unitValue;

  let roundedDateTime;
  if (unit === "week") {
    roundedDateTime = dateTime.startOf("day")[roundUp ? "plus" : "minus"]({
      day: roundUp ? 7 - smallestValue : smallestValue,
    });
  } else {
    roundedDateTime = dateTime
      .startOf(unit as DateTimeUnit)
      .set({ [unit]: roundedValue });
  }

  return roundedDateTime.toJSDate();
}

export function roundDownToTimeUnit(date, unit: DateTimeUnit | keyof DateTime) {
  const dateTime = DateTime.fromJSDate(date);
  if (!DateTime.isDateTime(dateTime)) {
    throw new Error("Invalid Luxon DateTime object");
  }

  return dateTime.startOf(unit as DateTimeUnit).toJSDate();
}
