import type { Axios } from "axios";

import { run as connection } from "./commands/connections.js";
import { run as backends } from "./commands/backends.js";
import { run as users } from "./commands/users.js";

export type PrintLine = (...str: any[]) => void;
export type KeyboardRead = (disableEcho?: boolean) => Promise<string>;

type Command = (
  args: string[],
  println: PrintLine,
  axios: Axios,
  apiKey: string,
  keyboardRead: KeyboardRead
) => Promise<void>;

type Commands = {
  name: string;
  description: string;
  run: Command;
}[];

export const commands: Commands = [
  {
    name: "help",
    description: "Prints help",
    async run(_args: string[], printf: PrintLine) {
      commands.forEach(command => {
        printf(`${command.name}: ${command.description}\n`);
      });

      printf("\nRun a command of your choosing with --help to see more options.\n");
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
    name: "conn",
    description: "Manages connections for NextNet",
    run: connection
  },
  {
    name: "user",
    description: "Manages users for NextNet",
    run: users
  },
  {
    name: "backend",
    description: "Manages backends for NextNet",
    run: backends
  },
  {
    name: "back",
    description: "(alias) Manages backends for NextNet",
    run: backends
  }
];
