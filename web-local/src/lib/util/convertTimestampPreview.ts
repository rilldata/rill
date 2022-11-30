import { removeTimezoneOffset as remove } from "./formatters";
/** used to convert a timestamp preview from the server for a sparkline. */
export function convertTimestampPreview(
  d,
  timeDimension: string,
  removeTimezoneOffset = false
) {
  return d.map((di) => {
    const pi = { ...di };
    pi.ts = removeTimezoneOffset
      ? remove(new Date(pi[timeDimension]))
      : new Date(pi[timeDimension]);

    delete pi[timeDimension];
    return pi;
  });
}
