import type { Axios } from "axios";

import { run as connection } from "./commands/connections.js";

export type PrintLine = (...str: string[]) => void;

type Command = (
  args: string[],
  println: PrintLine,
  axios: Axios,
  apiKey: string,
) => Promise<void>;

type Commands = {
  name: string;
  description: string;
  run: Command;
}[];

// TODO: add commands!

export const commands: Commands = [
  {
    name: "help",
    description: "Prints help",
    async run(_args: string[], printf: PrintLine) {
      commands.forEach(command => {
        printf(`${command.name}: ${command.description}\n`);
      });
    },
  },
  {
    name: "clear",
    description: "Clears screen",
    async run(_args: string[], printf: PrintLine) {
      printf("\x1B[2J\x1B[3J\x1B[H");
    },
  },
  {
    name: "connection",
    description: "Various connection related utilities",
    run: connection
  }
];
