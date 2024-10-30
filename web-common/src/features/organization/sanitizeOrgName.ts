const nameSanitiserRegex = /[^\w-]/g;

export function sanitizeOrgName(name: string) {
  return name.replace(nameSanitiserRegex, "-");
}
