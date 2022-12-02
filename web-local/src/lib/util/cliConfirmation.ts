import { createInterface } from "readline";

const DefaultConfirm = {
  yes: 1,
  y: 1,
};
const DefaultReject = {
  no: 1,
  n: 1,
};

export async function cliConfirmation(
  question: string,
  confirm = DefaultConfirm,
  reject = DefaultReject
): Promise<boolean> {
  const response = await getResponse(question);
  if (response.toLowerCase() in confirm) {
    return true;
  } else if (response.toLowerCase() in reject) {
    return false;
  } else {
    throw new Error("Invalid response");
  }
}

async function getResponse(question: string): Promise<string> {
  const rl = createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  return new Promise((resolve) => {
    rl.question(question + " ", async (answer) => {
      rl.close();

      resolve(answer.trim());
    });
  });
}
