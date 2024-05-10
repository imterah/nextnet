import { readFile, writeFile, mkdir } from "node:fs/promises";
import { format } from "node:util";

import parseArgsStringToArgv from "string-argv";
import baseAxios from "axios";
import ssh2 from "ssh2";

import { readFromKeyboard } from "./libs/readFromKeyboard.js";
import { commands } from "./commands.js";

let keyFile: Buffer | string | undefined;

const serverBaseURL: string =
  process.env.SERVER_BASE_URL ?? "http://127.0.0.1:3000/";

const axios = baseAxios.create({
  baseURL: serverBaseURL,
  validateStatus: () => true,
});

try {
  keyFile = await readFile("../keys/host.key");
} catch (e) {
  console.log("Error reading host key file! Creating new keypair...");
  await mkdir("../keys").catch(() => null);

  const keyPair: { private: string; public: string } = await new Promise(
    resolve =>
      ssh2.utils.generateKeyPair("ed25519", (err, keyPair) => resolve(keyPair)),
  );

  await writeFile("../keys/host.key", keyPair.private);
  await writeFile("../keys/host.pub", keyPair.public);

  keyFile = keyPair.private;
}

if (!keyFile) throw new Error("Somehow failed to fetch the key file!");

const server: ssh2.Server = new ssh2.Server({
  hostKeys: [keyFile],
  banner: "NextNet-LOM (c) NextNet project et al.",
});

server.on("connection", client => {
  let token: string = "";

  // eslint-disable-next-line
  let username: string = "";
  // eslint-disable-next-line
  let password: string = "";

  client.on("authentication", async auth => {
    if (auth.method == "password") {
      const response = await axios.post("/api/v1/users/login", {
        username: auth.username,
        password: auth.password,
      });

      if (response.status == 403) {
        return auth.reject(["password"]);
      }

      token = response.data.token;

      username = auth.username;
      password = auth.password;

      auth.accept();
    } else if (auth.method == "publickey") {
      return auth.reject();
      // todo
    } else {
      return auth.reject(["password", "publickey"]);
    }
  });

  client.on("ready", () => {
    client.on("session", accept => {
      const conn = accept();

      conn.on("exec", async (accept, reject, info) => {
        const stream = accept();

        // Matches on ; and &&
        const commandsRecv = info.command.split(/;|&&/).map(i => i.trim());

        function println(...data: unknown[]) {
          stream.write(format(...data).replaceAll("\n", "\r\n"));
        }

        for (const command of commandsRecv) {
          const argv = parseArgsStringToArgv(command);

          if (argv[0] == "exit") {
            stream.close();
          } else {
            const command = commands.find(i => i.name == argv[0]);

            if (!command) {
              stream.write(`Unknown command ${argv[0]}.\r\n`);

              continue;
            }

            await command.run(argv, println, axios, token, disableEcho =>
              readFromKeyboard(stream, disableEcho),
            );
          }
        }

        return stream.close();
      });

      // We're dumb. We don't really care.
      conn.on("pty", accept => accept());
      conn.on("window-change", accept => {
        if (typeof accept != "function") return;
        accept();
      });

      conn.on("shell", async accept => {
        const stream = accept();
        stream.write(
          "Welcome to NextNet LOM. Run 'help' to see commands.\r\n\r\n~$ ",
        );

        function println(...data: unknown[]) {
          stream.write(format(...data).replaceAll("\n", "\r\n"));
        }

        // FIXME (greysoh): wtf? this isn't setting correctly.
        // @eslint-disable-next-line
        while (true) {
          const line = await readFromKeyboard(stream);
          stream.write("\r\n");

          if (line == "") {
            stream.write(`~$ `);
            continue;
          }

          const argv = parseArgsStringToArgv(line);

          if (argv[0] == "exit") {
            stream.close();
          } else {
            const command = commands.find(i => i.name == argv[0]);

            if (!command) {
              stream.write(
                `Unknown command ${argv[0]}. Run 'help' to see commands.\r\n~$ `,
              );

              continue;
            }

            await command.run(argv, println, axios, token, disableEcho =>
              readFromKeyboard(stream, disableEcho),
            );
            stream.write("~$ ");
          }
        }
      });
    });
  });
});

server.listen(
  2222,
  process.env.NODE_ENV == "production" ? "0.0.0.0" : "127.0.0.1",
);

console.log("Started server at ::2222");
