import type { ServerChannel } from "ssh2";

const pullRate = process.env.KEYBOARD_PULLING_RATE ? parseInt(process.env.KEYBOARD_PULLING_RATE) : 5;

const leftEscape = "\x1B[D";
const rightEscape = "\x1B[C";

const ourBackspace = "\u0008";
const clientBackspace = "\x7F";

export async function readFromKeyboard(
  stream: ServerChannel,
  disableEcho: boolean = false,
): Promise<string> {
  let promise: (value: string | PromiseLike<string>) => void;

  let line = "";
  let lineIndex = 0;

  async function eventLoop(): Promise<any> {
    const readStreamData = stream.read();
    if (readStreamData == null) return setTimeout(eventLoop, pullRate);

    if (readStreamData.includes("\x03")) {
      stream.write("^C");
      return promise("");
    } else if (readStreamData.includes("\r") || readStreamData.includes("\n")) {
      return promise(line.replace("\r", ""));
    } else if (readStreamData.includes(clientBackspace)) {
      if (line.length == 0) return setTimeout(eventLoop, pullRate); // Here because if we do it in the parent if statement, shit breaks
      line = line.substring(0, lineIndex - 1) + line.substring(lineIndex);

      if (!disableEcho) {
        const deltaCursor = line.length - lineIndex;

        if (deltaCursor == line.length) return setTimeout(eventLoop, pullRate);

        if (deltaCursor < 0) {
          // Use old technique if the delta is < 0, as the new one is tailored to the start + 1 to end - 1
          stream.write(ourBackspace + " " + ourBackspace);
        } else {
          // Jump forward to the front, and remove the last character
          stream.write(rightEscape.repeat(deltaCursor) + " " + ourBackspace);

          // Go backwards & rerender text & go backwards again (wtf?)
          stream.write(
            leftEscape.repeat(deltaCursor + 1) +
              line.substring(lineIndex - 1) +
              leftEscape.repeat(deltaCursor + 1),
          );
        }

        lineIndex -= 1;
      }
    } else if (readStreamData.includes("\x1B")) {
      if (readStreamData.includes(rightEscape)) {
        if (lineIndex + 1 > line.length) return setTimeout(eventLoop, pullRate);
        lineIndex += 1;
      } else if (readStreamData.includes(leftEscape)) {
        if (lineIndex - 1 < 0) return setTimeout(eventLoop, pullRate);
        lineIndex -= 1;
      } else {
        return setTimeout(eventLoop, pullRate);
      }

      if (!disableEcho) stream.write(readStreamData);
    } else {
      lineIndex += readStreamData.length;

      // There isn't a splice method for String prototypes. So, ugh:
      line =
        line.substring(0, lineIndex - 1) +
        readStreamData +
        line.substring(lineIndex - 1);

      if (!disableEcho) {
        let deltaCursor = line.length - lineIndex;

        // wtf?
        if (deltaCursor < 0) {
          console.log(
            "FIXME: somehow, our deltaCursor value is negative! please investigate me",
          );
          deltaCursor = 0;
        }

        stream.write(
          line.substring(lineIndex - 1) + leftEscape.repeat(deltaCursor),
        );
      }
    }

    setTimeout(eventLoop, pullRate);
  };

  // Yes, this is bad practice. Currently, I don't care.
  return new Promise(async resolve => {
    setTimeout(eventLoop, pullRate);
    promise = resolve;
  });
}
