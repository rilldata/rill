import dotenv from "dotenv";
import path from "node:path";
import { fileURLToPath } from "node:url";

/**
 * Users of the embed host, keyed by the `?user=` param.
 * Their emails are suffixed variants of the e2e admin's email, so Mailpit catches them,
 * but they don't match any project member.
 */
export type EmbedUserKey = "none" | "alice" | "bob" | "carol" | "admin";

export type EmbedUser = {
  key: EmbedUserKey;
  label: string;
  /** The email of the user; also used as the external user ID. Empty for the anonymous user. */
  email: string;
  /** Passes the email as `user_email` to `GetIFrame`, which prefills and locks the creator email in the report form. */
  passUserEmail?: boolean;
  attributes?: Record<string, unknown>;
};

let adminEmail: string | undefined;
let users: EmbedUser[] | undefined;

export function getEmbedUsers(): EmbedUser[] {
  if (users) return users;

  const alice = withSuffix(getAdminEmail(), "embed-a");
  const bob = withSuffix(getAdminEmail(), "embed-b");
  const carol = withSuffix(getAdminEmail(), "embed-c");
  const admin = withSuffix(getAdminEmail(), "embed-admin");
  users = [
    { key: "none", label: "No user", email: "" },
    {
      key: "alice",
      label: "Alice",
      email: alice,
      attributes: { email: alice },
    },
    { key: "bob", label: "Bob", email: bob, attributes: { email: bob } },
    { key: "carol", label: "Carol", email: carol, passUserEmail: true },
    {
      key: "admin",
      label: "Admin",
      email: admin,
      attributes: { email: admin, admin: true },
    },
  ];
  return users;
}

export function getEmbedUser(key: EmbedUserKey): EmbedUser {
  const user = getEmbedUsers().find((u) => u.key === key);
  if (!user) throw new Error(`Unknown embed user "${key}"`);
  return user;
}

/** An address that isn't an embed user, used as an extra report recipient. */
export function getExtraRecipientEmail() {
  return withSuffix(getAdminEmail(), "embed-extra");
}

function getAdminEmail() {
  if (adminEmail) return adminEmail;

  // e2e workers don't load the root `.env` (only the setup project does), so load it here.
  // It's loaded lazily, so specs that don't use the embed host don't need it.
  const repoRoot = path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    "../../../",
  );
  dotenv.config({ path: path.join(repoRoot, ".env") });
  adminEmail = process.env.RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL;
  if (!adminEmail) {
    throw new Error(
      "Missing RILL_DEVTOOL_E2E_ADMIN_ACCOUNT_EMAIL; run `npm run test:setup` to pull the root .env file",
    );
  }
  return adminEmail;
}

/** Inserts `+suffix` before the `@`; for example, `e2e-admin@rilldata.com` becomes `e2e-admin+suffix@rilldata.com`. */
function withSuffix(email: string, suffix: string) {
  const at = email.lastIndexOf("@");
  return `${email.slice(0, at)}+${suffix}${email.slice(at)}`;
}
