// Letters and digits that are easy to read aloud and transcribe: no 0/O, 1/l/I.
const ALPHABET = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789";

export const GENERATED_PASSWORD_LENGTH = 16;

/** Generates a random password for a public URL using the platform's CSPRNG. */
export function generatePassword(length = GENERATED_PASSWORD_LENGTH): string {
  // Reject bytes above the largest multiple of the alphabet size to avoid modulo bias.
  const limit = 256 - (256 % ALPHABET.length);
  let password = "";
  while (password.length < length) {
    const bytes = crypto.getRandomValues(new Uint8Array(length));
    for (const byte of bytes) {
      if (byte >= limit) continue;
      password += ALPHABET[byte % ALPHABET.length];
      if (password.length === length) break;
    }
  }
  return password;
}
