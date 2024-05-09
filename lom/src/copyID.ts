import { writeFile } from "node:fs/promises";
import ssh2 from "ssh2";

import { readFromKeyboard } from "./libs/readFromKeyboard.js";
import type { ClientKeys } from "./index.js";

export async function runCopyID(username: string, password: string, keys: ClientKeys, stream: ssh2.ServerChannel) {
  stream.write("Hey there! I think you're using ssh-copy-id. If this is an error, you may close this terminal.\n");
  stream.write("Please wait...\n");

  const keyData = await readFromKeyboard(stream, true);
  stream.write("Parsing key...\n");

  const parsedKey = ssh2.utils.parseKey(keyData);
  
  if (parsedKey instanceof Error) {
    stream.write(parsedKey.message + "\n");
    return stream.close();
  }

  stream.write("Passed checks. Writing changes...\n");
  
  keys.push({
    username,
    password,
    publicKey: keyData
  });

  try {
    await writeFile("../keys/clients.json", JSON.stringify(keys, null, 2));
  } catch (e) {
    console.log(e);
    return stream.write("ERROR: Failed to save changes! If you're the administrator, view the console for details.\n");
  }

  stream.write("Success!\n");
  return stream.close();
}