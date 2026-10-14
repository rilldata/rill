export function isValidPublicURLPassword(password: string): boolean {
  return (
    Array.from(password).length >= 8 &&
    new TextEncoder().encode(password).length <= 72
  );
}
