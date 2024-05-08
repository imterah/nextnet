import type { Axios } from "axios";

import { SSHCommand } from "../libs/patchCommander.js";
import type { PrintLine, KeyboardRead } from "../commands.js";

type UserLookupSuccess = {
  success: true;
  data: {
    id: number,
    isServiceAccount: boolean,
    username: string,
    name: string,
    email: string
  }[];
};

export async function run(
  argv: string[],
  println: PrintLine,
  axios: Axios,
  apiKey: string,
  readKeyboard: KeyboardRead
) {
  const program = new SSHCommand(println);
  program.description("Manages users for NextNet");
  program.version("v1.0.0-testing");

  const addCommand = new SSHCommand(println, "add");
  addCommand.description("Create a new user");
  addCommand.argument("<username>", "Username of new user");
  addCommand.argument("<email>", "Email of new user");
  addCommand.argument("[name]", "Name of new user (defaults to username)");

  addCommand.option("-p, --password", "Password of User");
  addCommand.option(
    "-a, --ask-password, --ask-pass, --askpass",
    "Asks for a password. Hides output",
  );

  addCommand.action(async(username: string, email: string, name: string, options: {
    password?: string,
    askPassword?: boolean
  }) => {
    if (!options.password && !options.askPassword) {
      println("No password supplied, and askpass has not been supplied.\n");
      return;
    };

    let password: string = "";

    if (options.askPassword) {
      let passwordConfirmOne = "a";
      let passwordConfirmTwo = "b";

      while (passwordConfirmOne != passwordConfirmTwo) {
        println("Password: ");
        passwordConfirmOne = await readKeyboard(true);
        
        println("\nConfirm password: ");
        passwordConfirmTwo = await readKeyboard(true);

        println("\n");

        if (passwordConfirmOne != passwordConfirmTwo) {
          println("Passwords do not match! Try again.\n\n");
        }
      }

      password = passwordConfirmOne;
    } else {
      // From the first check we do, we know this is safe (you MUST specify a password)
      // @ts-ignore
      password = options.password;
    }

    const response = await axios.post("/api/v1/users/create", {
      name,
      username,
      email,
      password
    });

    if (response.status != 200) {
      if (process.env.NODE_ENV != "production") console.log(response);

      if (response.data.error) {
        println(`Error: ${response.data.error}\n`);
      } else {
        println("Error requesting connections!\n");
      }

      return;
    }

    println("User created successfully.\n");
  })

  const removeCommand = new SSHCommand(println, "rm");
  removeCommand.description("Remove a user");
  removeCommand.argument("<uid>", "ID of user to remove");

  removeCommand.action(async(uidStr: string) => {
    const uid = parseInt(uidStr);

    if (Number.isNaN(uid)) {
      println("UID (%s) is not a number.\n", uid);
      return;
    }

    let response = await axios.post("/api/v1/users/remove", {
      token: apiKey,
      uid
    });

    if (response.status != 200) {
      if (process.env.NODE_ENV != "production") console.log(response);

      if (response.data.error) {
        println(`Error: ${response.data.error}\n`);
      } else {
        println("Error deleting user!\n");
      }

      return;
    }
    
    println("User has been successfully deleted.\n");
  });

  const lookupCommand = new SSHCommand(println, "find");
  lookupCommand.description("Find a user");
  lookupCommand.option("-i, --id <id>", "UID of User");
  lookupCommand.option("-n, --name <name>", "Name of User");
  lookupCommand.option("-u, --username <username>", "Username of User");
  lookupCommand.option("-e, --email <email>", "Email of User");
  lookupCommand.option("-s, --service", "The user is a service account");

  lookupCommand.action(async(options) => {
    // FIXME: redundant parseInt calls

    if (options.id) {
      const uid = parseInt(options.id);

      if (Number.isNaN(uid)) {
        println("UID (%s) is not a number.\n", uid);
        return;
      };
    }

    const response = await axios.post("/api/v1/users/lookup", {
      token: apiKey,
      id: options.id ? parseInt(options.id) : undefined,
      name: options.name,
      username: options.username,
      email: options.email,
      service: Boolean(options.service)
    });

    if (response.status != 200) {
      if (process.env.NODE_ENV != "production") console.log(response);

      if (response.data.error) {
        println(`Error: ${response.data.error}\n`);
      } else {
        println("Error finding user!\n");
      }

      return;
    }

    const { data }: UserLookupSuccess = response.data;

    data.forEach((user) => {
      println("UID: %s%s:\n", user.id, (user.isServiceAccount ? " (service)" : ""));
      println("- Username: %s\n", user.username);
      println("- Name: %s\n", user.name);
      println("- Email: %s\n", user.email);

      println("\n");
    });

    println("%s users found.\n", response.data.data.length);
  })

  program.addCommand(addCommand);
  program.addCommand(removeCommand);
  program.addCommand(lookupCommand);

  program.parse(argv);
  await new Promise((resolve) => program.onExit(resolve));
}
