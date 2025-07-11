export function isEmpty(val: any) {
  return (
    val === undefined ||
    val === null ||
    val === "" ||
    (typeof val === "string" && val.trim() === "")
  );
}

export function normalizeErrors(
  err: any,
): string | string[] | null | undefined {
  if (!err) return undefined;
  if (Array.isArray(err)) return err;
  if (typeof err === "string") return err;
  if (err._errors && Array.isArray(err._errors)) return err._errors;
  return undefined;
}
