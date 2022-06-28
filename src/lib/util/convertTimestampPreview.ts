/** used to convert a timestamp preview from the server for a sparkline. */
export function convertTimestampPreview(d) {
  return d.map((di) => {
    let pi = { ...di };
    pi.ts = new Date(pi.ts);
    return pi;
  });
}
