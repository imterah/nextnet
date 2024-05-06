import { readFile, writeFile, mkdir } from "node:fs/promises";

import parseArgsStringToArgv from "string-argv";
import baseAxios from "axios";
import ssh2 from "ssh2";

import { commands } from "./commands.js";

let keyFile: Buffer | string | undefined;

const serverBaseURL: string =
  process.env.SERVER_BASE_URL ?? "http://127.0.0.1:3000/";

const axios = baseAxios.create({
  baseURL: serverBaseURL,
  validateStatus: () => true
});

async function readFromKeyboard(stream: ssh2.ServerChannel, disableEcho: boolean = false): Promise<string> {
  const leftEscape = "\x1B[D";
  const rightEscape = "\x1B[C";
  
  let line = "";
  let lineIndex = 0;
  let isReady = false;

  async function eventLoop(): Promise<any> {
    const readStreamData = stream.read();
    if (readStreamData == null) return setTimeout(eventLoop, 5);

    if (readStreamData.includes("\r") || readStreamData.includes("\n")) {
      line = line.replace("\r", "");
      isReady = true;

      return;
    } else if (readStreamData.includes("\x7F")) {
      // \x7F = Ascii escape code for backspace   (client side)
      // \u0008 = Ascii escape code for backspace (server side)
      line = line.substring(0, line.length - 1);
      if (!disableEcho) stream.write("\u0008 \u0008");
    } else if (readStreamData.includes("\x1B")) {    
      if (readStreamData.includes(rightEscape)) {
        if (lineIndex + 1 > line.length) return setTimeout(eventLoop, 5);
        lineIndex += 1;
      } else if (readStreamData.includes(leftEscape)) {
        if (lineIndex - 1 < 0) return setTimeout(eventLoop, 5);
        lineIndex -= 1;
      } else {
        return setTimeout(eventLoop, 5);
      }

      if (!disableEcho) stream.write(readStreamData);
    } else {
      lineIndex += readStreamData.length;

      // There isn't a splice method for String prototypes. So, ugh:
      line = line.substring(0, lineIndex - 1) + readStreamData + line.substring(lineIndex - 1);
      
      if (!disableEcho) {
        let deltaCursor = line.length - lineIndex;
      
        // wtf?
        if (deltaCursor < 0) {
          console.log("FIXME: somehow, our deltaCursor value is negative! please investigate me");
          deltaCursor = 0;
        }
      
        stream.write(line.substring(lineIndex - 1) + leftEscape.repeat(deltaCursor));
      }
    }

    setTimeout(eventLoop, 5);
  }
  
  // Yes, this is bad practice. Currently, I don't care.
  return new Promise(async(resolve) => {
    eventLoop();

    while (!isReady) {
      await new Promise((i) => setTimeout(i, 5));
    }

    resolve(line);
  });
};

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
          stream.write(str.join(" ").replace("\n", "\r\n"));
        };

        while (true) {
          const line = await readFromKeyboard(stream);
          stream.write("\r\n");
          
          if (line == "") continue;
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
