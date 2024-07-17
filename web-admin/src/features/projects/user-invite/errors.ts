import type { AxiosError } from "axios";

export function parseError(error: AxiosError, email: string) {
  return `${email}: ${error.response?.data?.message ?? error.message}`;
}
