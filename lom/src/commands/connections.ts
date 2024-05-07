import { Command } from "commander";
import type { Axios } from "axios";

import { type AppState, patchCommander } from "../libs/patchCommander.js";
import type { PrintLine } from "../commands.js";

export async function run(
  argv: string[],
  println: PrintLine,
  axios: Axios,
  apiKey: string,
) {
  const program = new Command();
  const appState: AppState = {
    hasRecievedExitSignal: false
  };

  patchCommander(program, appState, println);

  program
    .option('-d, --debug', 'output extra debugging')
    .option('-s, --small', 'small pizza size')
    .option('-p, --pizza-type <type>', 'flavour of pizza');

  program.parse(["node", ...argv]);

  if (appState.hasRecievedExitSignal) return;
}
