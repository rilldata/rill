export type Person = {
  firstName: string;
  lastName: string;
  age: number;
  visits: number;
  progress: number;
  status: "relationship" | "complicated" | "single";
  subRows?: Person[];
};

const range = (len: number) => {
  const arr = [];
  for (let i = 0; i < len; i++) {
    arr.push(i);
  }
  return arr;
};

function getRandomInt(max) {
  return Math.floor(Math.random() * Math.floor(max));
}

function getRandomElement(array) {
  return array[Math.floor(Math.random() * array.length)];
}

const newPerson = () => {
  const firstNames = ["John", "Jane", "Alice", "Bob", "Chris", "Sara"];
  const lastNames = ["Doe", "Smith", "Johnson", "Williams", "Brown", "Jones"];

  return {
    firstName: getRandomElement(firstNames),
    lastName: getRandomElement(lastNames),
    age: getRandomInt(40),
    visits: getRandomInt(1000),
    progress: getRandomInt(100),
    status: getRandomElement(["relationship", "complicated", "single"]),
  };
};

export function makeData(...lens: number[]) {
  const makeDataLevel = (depth = 0): Person[] => {
    const len = lens[depth]!;
    return range(len).map((d): Person => {
      return {
        ...newPerson(),
        subRows: lens[depth + 1] ? makeDataLevel(depth + 1) : undefined,
      };
    });
  };

  return makeDataLevel();
}
