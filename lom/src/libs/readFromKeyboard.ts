import type { ServerChannel } from "ssh2";

export async function readFromKeyboard(stream: ServerChannel, disableEcho: boolean = false): Promise<string> {
  const leftEscape = "\x1B[D";
  const rightEscape = "\x1B[C";

  const ourBackspace = "\u0008";

  // \x7F = Ascii escape code for backspace   (client side)
  
  let line = "";
  let lineIndex = 0;
  let isReady = false;

  async function eventLoop(): Promise<any> {
    const readStreamData = stream.read();
    if (readStreamData == null) return setTimeout(eventLoop, 5);

    if (readStreamData.includes("\r") || readStreamData.includes("\n")) {
      line = line.replace("\r", "");
      isReady = true;

      return;
    } else if (readStreamData.includes("\x7F")) {
      // TODO (greysoh): investigate maybe potential deltaCursor overflow (haven't tested)
      // TODO (greysoh): investigate deltaCursor underflow
      // TODO (greysoh): make it not run like shit (I don't know how to describe it)
      
      if (line.length == 0) return setTimeout(eventLoop, 5); // Here because if we do it in the parent if statement, shit breaks
      line = line.substring(0, lineIndex - 1) + line.substring(lineIndex);

      if (!disableEcho) {
        let deltaCursor = line.length - lineIndex;

        // wtf?
        if (deltaCursor < 0) {
          console.log("FIXME: somehow, our deltaCursor value is negative! please investigate me");
          return setTimeout(eventLoop, 5);
        }

        // Jump forward to the front, and remove the last character
        stream.write(rightEscape.repeat(deltaCursor) + " " + ourBackspace);

        // Go backwards & rerender text & go backwards again (wtf?)
        stream.write(leftEscape.repeat(deltaCursor + 1) + line.substring(lineIndex - 1) + leftEscape.repeat(deltaCursor + 1));
      }
    } else if (readStreamData.includes("\x1B")) {    
      if (readStreamData.includes(rightEscape)) {
        if (lineIndex + 1 > line.length) return setTimeout(eventLoop, 5);
        lineIndex += 1;
      } else if (readStreamData.includes(leftEscape)) {
        if (lineIndex - 1 < 0) return setTimeout(eventLoop, 5);
        lineIndex -= 1;
      } else {
        return setTimeout(eventLoop, 5);
      }

      if (!disableEcho) stream.write(readStreamData);
    } else {
      lineIndex += readStreamData.length;

      // There isn't a splice method for String prototypes. So, ugh:
      line = line.substring(0, lineIndex - 1) + readStreamData + line.substring(lineIndex - 1);
      
      if (!disableEcho) {
        let deltaCursor = line.length - lineIndex;
      
        // wtf?
        if (deltaCursor < 0) {
          console.log("FIXME: somehow, our deltaCursor value is negative! please investigate me");
          deltaCursor = 0;
        }
      
        stream.write(line.substring(lineIndex - 1) + leftEscape.repeat(deltaCursor));
      }
    }

    setTimeout(eventLoop, 5);
  }
  
  // Yes, this is bad practice. Currently, I don't care.
  return new Promise(async(resolve) => {
    eventLoop();

    while (!isReady) {
      await new Promise((i) => setTimeout(i, 5));
    }

    resolve(line);
  });
};