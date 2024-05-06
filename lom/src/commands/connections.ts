import { parseArgs } from "node:util";

import type { Axios } from "axios";
import type { PrintLine } from "../commands.js";

export async function run(
  args: string[],
  println: PrintLine,
  axios: Axios,
  apiKey: string,
) {
  const options = parseArgs({
    args,

    strict: false,
    allowPositionals: true,

    options: {
      tail: {
        type: "boolean",
        short: "t",
        default: false,
      },
    },
  });

  // Special filtering
  const values = options.values;
  const positionals = options.positionals
    .map(i => (!i.startsWith("-") ? i : ""))
    .filter(Boolean);
}
