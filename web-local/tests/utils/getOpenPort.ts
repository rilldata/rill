import { createServer } from "net";

export async function getOpenPort(): Promise<number> {
  return new Promise((res) => {
    const srv = createServer();
    srv.listen(0, () => {
      const address = srv?.address();
      if (!address || typeof address === "string") {
        throw new Error("Invalid address");
      }
      srv.close(() => res(address.port));
    });
  });
}
