function getRandomInt(min: number, max: number): number {
  const minCeiled = Math.ceil(min);
  const maxFloored = Math.floor(max);
  return Math.floor(Math.random() * (maxFloored - minCeiled) + minCeiled); // The maximum is exclusive and the minimum is inclusive
}

export function generateRandomData(length: number = 128): string {
  let newString = "";

  for (let i = 0; i < length; i += 2) {
    const randomNumber = getRandomInt(0, 255);

    if (randomNumber == 0) {
      i -= 2;
      continue;
    }

    newString += randomNumber.toString(16);
  }

  return newString;
}
