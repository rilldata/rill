import axios from "axios";
import { expect } from "@playwright/test";

const MAIL_PIT_API_URL = "http://localhost:8025/api/v1";
type MailPitEmail = {
  ID: string;
  Subject: string;
  Created: string;
};
type MailPitCompleteEmail = MailPitEmail & {
  Text: string;
};
type MailPitMessagesResponse = {
  messages: MailPitEmail[];
};

export async function waitForEmail(title: string, after: Date) {
  let email: MailPitEmail | undefined = undefined;
  await expect
    .poll(
      async () => {
        email = await getEmail(title, after);
        return !!email;
      },
      {
        intervals: Array(10).fill(2_000),
        timeout: 20_000,
        message: `Email with title "${title}" not found`,
      },
    )
    .toBeTruthy();
  return email!;
}

const LinkExtractorRegex = /Open in browser \(\s*(.*?)\s*\)/;
export async function getOpenLinkFromEmail(email: MailPitEmail) {
  try {
    const resp = await axios.get(`${MAIL_PIT_API_URL}/message/${email.ID}`);
    const completeEmail = resp.data as MailPitCompleteEmail;
    const link = LinkExtractorRegex.exec(completeEmail.Text)?.[1];
    if (!link)
      throw new Error(`No link found in email with subject "${email.Subject}"`);
    return decodeURI(link);
  } catch (err) {
    console.error(err.errors?.[0]);
    throw new Error(
      `Error fetching email with subject "${email.Subject}". Ensure if mailpit is running in docker.`,
    );
  }
}

async function getEmail(title: string, after: Date) {
  try {
    const resp = await axios.get(`${MAIL_PIT_API_URL}/messages`);
    const messages = (resp.data as MailPitMessagesResponse).messages;
    return messages.find(
      (m) =>
        m.Subject.includes(title) &&
        new Date(m.Created).getTime() >= after.getTime(),
    );
  } catch (err) {
    console.error(err.errors?.[0]);
    throw new Error(
      `Error fetching email with subject "${title}". Ensure if mailpit is running in docker.`,
    );
  }
}
