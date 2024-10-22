export function getSingleFieldError(fieldLabel: string, field: string) {
  return new Error(`Select ${fieldLabel}: "${field}" is not valid.`);
}

export function getMultiFieldError(fieldLabel: string, fields: string[]) {
  const plural = fields.length > 1;
  return new Error(
    `Select ${fieldLabel}${plural ? "s" : ""}: "${fields.join(",")}" ${plural ? "are" : "is"} not valid.`,
  );
}
