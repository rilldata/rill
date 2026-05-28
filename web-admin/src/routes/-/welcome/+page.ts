import { redirect } from "@sveltejs/kit";

export const load = () => {
  // Safeguard against direct access to /-/welcome
  throw redirect(307, "/-/welcome/theme");
};
