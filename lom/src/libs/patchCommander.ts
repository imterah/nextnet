import { Command, type ParseOptions } from "commander";
import { PrintLine } from "../commands";

export class SSHCommand extends Command {
  hasRecievedExitSignal: boolean;
  println: PrintLine;

  /**
   * Modified version of the Commander command with slight automated patches, to work with our SSH environment.
   * @param println PrintLine function to use
   * @param name Optional field for the name of the command
   */
  constructor(
    println: PrintLine,
    name?: string,
    disableSSHHelpPatching: boolean = false,
  ) {
    super(name);

    this.configureOutput({
      writeOut: str => println(str),
      writeErr: str => {
        if (str.includes("--help") || str.includes("-h")) return;
        println(str);
      },
    });

    if (!disableSSHHelpPatching) {
      const sshCommand = new SSHCommand(println, "help", true);

      sshCommand.description("display help for command");
      sshCommand.argument("[command]", "command to show help for");
      sshCommand.action(() => {
        this.hasRecievedExitSignal = true;

        if (process.env.NODE_ENV != "production") {
          println(
            "Caught irrecoverable action (command help call) in patchCommander\n",
          );
        } else {
          println("Aborted\n");
        }
      });

      this.addCommand(sshCommand);
    }
  }

  _exit() {
    this.hasRecievedExitSignal = true;
  }

  _exitCallback() {
    this.hasRecievedExitSignal = true;
  }

  action(fn: (...args: any[]) => void | Promise<void>): this {
    super.action(fn);

    // @ts-ignore
    // prettier-ignore
    const oldActionHandler: (...args: any[]) => void | Promise<void> = this._actionHandler;

    // @ts-ignore
    this._actionHandler = async (...args: any[]): Promise<void> => {
      if (args[0][0] == "--help" || args[0][0] == "-h") return;
      await oldActionHandler(...args);
    };

    return this;
  }

  parse(argv?: readonly string[], options?: ParseOptions): this {
    super.parse(["nextruntime", ...(argv ?? [])], options);
    return this;
  }

  createCommand(name: string) {
    const command = new SSHCommand(this.println, name);
    return command;
  }
}
