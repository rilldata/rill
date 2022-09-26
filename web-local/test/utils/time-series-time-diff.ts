import { MICROS } from "@rilldata/web-local/common/database-service/DatabaseColumnActions";

const MinMonth = (MICROS.day / 1000) * 28;
const MaxMonth = (MICROS.day / 1000) * 31;

export function isTimestampDiffAccurate(
  fromTimestamp: string,
  toTimestamp: string,
  rollupInterval: string
): boolean {
  if (rollupInterval === "year") {
    throw new Error("Unsupported interval");
  }
  const fromDate = new Date(fromTimestamp);
  const toData = new Date(toTimestamp);
  const dateDiff = toData.getTime() - fromDate.getTime();

  if (rollupInterval === "month") {
    return dateDiff >= MinMonth && dateDiff <= MaxMonth;
  } else {
    return dateDiff === MICROS[rollupInterval] / 1000;
  }
}
