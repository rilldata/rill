import {createServer} from "net";

export async function isPortOpen(port: number): Promise<boolean> {
    return new Promise((resolve, reject) => {
        const testServer = createServer();
        testServer
            .once("error", (err: any) => {
                if (err.code != "EADDRINUSE") reject(err);
                else resolve(true);
            })
            .once("listening", () => {
                testServer.close(() => {
                    resolve(false);
                });
            });
        testServer.listen(port);
    });
}
