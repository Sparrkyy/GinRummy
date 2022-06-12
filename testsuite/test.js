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
        myID = parseInt(data["content"])
      }
      if (
        data["messagetype"] === "meta" &&
        data["command"] === "gameroomstatus" &&
        data["content"] === "filled"
      ) {
        const gamedata = data.game;
        if (gamedata.deck.length === 31 && gamedata.player1hand.length === 10 && gamedata.player2hand.length === 10){
          console.log("Passed: All lenghts are correct")
          console.log("MyID:", myID)
          console.log("player1 info", gamedata.player1)
          console.log("player2 info", gamedata.player2)
          console.log("the players turn:", gamedata.turn)
        }
        else {
          console.error("Failed: not the right number of cards", gamedata.deck.length, gamedata.player1hand.length, gamedata.player2hand.length)
        }
        if (gamedata.player1.id === myID && gamedata.turn === myID) {
          console.log("Passed: my id is the same as the turn and the same as player1");
          resolve()
        } else {
          console.error("Failed: not proper result", gamedata.player1.id, myID, gamedata.turn)
          resolve()
        }
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
