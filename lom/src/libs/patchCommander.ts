// @ts-nocheck

import type { Command } from "commander";
import { PrintLine } from "../commands";

export type AppState = {
  hasRecievedExitSignal: boolean
}

export function patchCommander(program: Command, appState: AppState, println: PrintLine) {
  program.exitOverride(() => {
    appState.hasRecievedExitSignal = true;
  });

  program._outputConfiguration.writeOut = (str) => println(str);
  program._outputConfiguration.writeErr = (str) => {
    if (str.includes("--help")) return;
    println(str);
  };

  program._exit = (exitCode, code, message) => {
    appState.hasRecievedExitSignal = true;
  }
}