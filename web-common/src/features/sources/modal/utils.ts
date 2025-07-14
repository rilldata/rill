export function isEmpty(val: any) {
  return (
    val === undefined ||
    val === null ||
    val === "" ||
    (typeof val === "string" && val.trim() === "")
  );
}
