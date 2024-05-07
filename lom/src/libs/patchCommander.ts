// @ts-nocheck

import { Command } from "commander";
import { PrintLine } from "../commands";

export type AppState = {
  hasRecievedExitSignal: boolean
}

export function patchCommander(program: Command, appState: AppState, println: PrintLine) {
  program._outputConfiguration.writeOut = (str) => println(str);
  program._outputConfiguration.writeErr = (str) => {
    if (str.includes("--help")) return;
    println(str);
  };

  program._exit = () => {
    appState.hasRecievedExitSignal = true;
  };

  program.createCommand = (name: string) => {
    const command = new Command(name);
    patchCommander(command, appState, println);

    return command;
  };

  program._exitCallback = () => {
    appState.hasRecievedExitSignal = true;
  };
}