let TEST_COUNT = 1;
let WS1;
let WS2;

const waitTillConected = (websocket) => {
  return new Promise((resolve) => {
    websocket.onopen = () => {
      resolve();
    };
  });
};

const print = (string) => {
  console.log(string);
};

const test = async (discription, test) => {
  const message = "# Test " + TEST_COUNT + ": " + discription + " #";
  TEST_COUNT++;
  let barrier = "";
  for (let i = 0; i < message.length; i++) {
    barrier += "#";
  }
  print(barrier);
  print(message);
  print(barrier);
  await test();
};

const test1_BasicConnect = async () => {
  return new Promise((resolve) => {
    try {
      WS1 = new WebSocket("ws://localhost:8080/channel/test1/play");
      WS2 = new WebSocket("ws://localhost:8080/channel/test1/play");
      WS1.onopen = () => {
        console.log("Passed");
        resolve();
      };
    } catch (e) {
      console.error("Failed", e);
      resolve();
    }
  });
};

const test2_figureOutMetaData = async () => {
  return new Promise((resolve) => {
    WS1 = new WebSocket("ws://localhost:8080/channel/test2/play");
    WS2 = new WebSocket("ws://localhost:8080/channel/test2/play");
    WS1.onmessage = (messageIn) => {
      const data = JSON.parse(messageIn.data);
      if (data["messagetype"] !== undefined && data["messagetype"] === "meta") {
        if (data["command"] === "opponent") {
          console.log("Passed: player1's opponent", data["content"]);
          resolve();
        }
      }
    };
  });
};

const test3_figureoutwhosturn = async () => {
  return new Promise((resolve) => {
    WS1 = new WebSocket("ws://localhost:8080/channel/test3/play");
    WS2 = new WebSocket("ws://localhost:8080/channel/test3/play");
    let myID = null;

    WS1.onmessage = (messageIn) => {
      let data = JSON.parse(messageIn.data);
      if (data["messagetype"] === "meta" && data["command"] === "iam") {
        myID = data["content"];
      }
      if (
        data["messagetype"] === "meta" &&
        data["command"] === "gameroomstatus" &&
        data["content"] === "filled"
      ) {
        console.log(data);
      }
    };
  });
};

const runtests = async () => {
  await test("Basic Connection", test1_BasicConnect);
  await test("finding the oppenet", test2_figureOutMetaData);
  await test("finding the intial game state", test3_figureoutwhosturn);
};

runtests();
