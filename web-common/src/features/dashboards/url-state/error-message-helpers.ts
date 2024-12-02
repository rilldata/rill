export function getSingleFieldError(fieldLabel: string, field: string) {
  return new Error(`Selected ${fieldLabel}: "${field}" is not valid.`);
}

export function getMultiFieldError(fieldLabel: string, fields: string[]) {
  const plural = fields.length > 1;
  return new Error(
    `Selected ${fieldLabel}${plural ? "s" : ""}: "${fields.join(",")}" ${plural ? "are" : "is"} not valid.`,
  );
}
