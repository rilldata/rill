import axios from "axios";
import { expect } from "@playwright/test";

const MAIL_PIT_API_URL = "http://localhost:8025/api/v1";
type MailPitAddress = {
  Address: string;
};
type MailPitEmail = {
  ID: string;
  Subject: string;
  Created: string;
  To: MailPitAddress[];
};
export type MailPitCompleteEmail = MailPitEmail & {
  Text: string;
  Attachments: { FileName: string }[];
};
type MailPitMessagesResponse = {
  messages: MailPitEmail[];
};

/**
 * Waits for an email with the title, created after the given time.
 * If `to` is set, only matches emails sent to that address.
 */
export async function waitForEmail(title: string, after: Date, to?: string) {
  let email: MailPitEmail | undefined = undefined;
  await expect
    .poll(
      async () => {
        email = await getEmail(title, after, to);
        return !!email;
      },
      {
        intervals: Array(10).fill(2_000),
        timeout: 20_000,
        message: `Email with title "${title}"${to ? ` to ${to}` : ""} not found`,
      },
    )
    .toBeTruthy();
  return email!;
}

const LinkExtractorRegex = /Open in browser \(\s*(.*?)\s*\)/;
export async function getOpenLinkFromEmail(email: MailPitEmail) {
  const completeEmail = await getCompleteEmail(email);
  const link = extractOpenLink(completeEmail);
  if (!link)
    throw new Error(`No link found in email with subject "${email.Subject}"`);
  return link;
}

/** Returns the open link in the email's text, or undefined if it has none. */
export function extractOpenLink(email: MailPitCompleteEmail) {
  const link = LinkExtractorRegex.exec(email.Text)?.[1];
  return link ? decodeURI(link) : undefined;
}

/** Fetches the full email, including its text and attachments. */
export async function getCompleteEmail(email: MailPitEmail) {
  try {
    const resp = await axios.get(`${MAIL_PIT_API_URL}/message/${email.ID}`);
    return resp.data as MailPitCompleteEmail;
  } catch (err) {
    console.error(err.errors?.[0]);
    throw new Error(
      `Error fetching email with subject "${email.Subject}". Ensure if mailpit is running in docker.`,
    );
  }
}

async function getEmail(title: string, after: Date, to?: string) {
  try {
    const resp = await axios.get(`${MAIL_PIT_API_URL}/messages`);
    const messages = (resp.data as MailPitMessagesResponse).messages;
    return messages.find(
      (m) =>
        m.Subject.includes(title) &&
        new Date(m.Created).getTime() >= after.getTime() &&
        (!to || m.To.some((r) => r.Address.toLowerCase() === to.toLowerCase())),
    );
  } catch (err) {
    console.error(err.errors?.[0]);
    throw new Error(
      `Error fetching email with subject "${title}". Ensure if mailpit is running in docker.`,
    );
  }
}
