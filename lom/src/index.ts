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
  validateStatus: () => true
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
  greeting: "NextNet LOM (beta)",
});

server.on("connection", client => {
  let token: string = "";

  client.on("authentication", async auth => {
    if (auth.method != "password") return auth.reject(["password"]); // We need to know the password to auth with the API

    const response = await axios.post("/api/v1/users/login", {
      username: auth.username,
      password: auth.password,
    });

    if (response.status == 403) {
      return auth.reject(["password"]);
    }

    token = response.data.token;
    auth.accept();
  });

  client.on("ready", () => {
    client.on("session", (accept, reject) => {
      const conn = accept();

      // We're dumb. We don't really care.
      conn.on("pty", accept => accept());
      conn.on("window-change", accept => accept());

      conn.on("shell", async accept => {
        const stream = accept();
        stream.write(
          "Welcome to NextNet LOM. Run 'help' to see commands.\r\n\r\n~$ ",
        );

        function println(...str: string[]) {
          stream.write(format(...str).replace("\n", "\r\n"));
        };

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
    
            await command.run(argv, println, axios, token);
            stream.write(`~$ `);
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
