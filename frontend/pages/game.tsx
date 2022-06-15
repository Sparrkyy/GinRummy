import type { NextPage } from "next";
import AuthContext from "./AuthContext";
import { useRouter } from "next/router";
import { useContext, useEffect, useState, useRef } from "react";
import Button from "react-bootstrap/button";
import Form from "react-bootstrap/Form";
import Alert from "react-bootstrap/Alert"
import "bootstrap/dist/css/bootstrap.min.css";
import { w3cwebsocket as W3CWebSocket } from "websocket";

interface websocket {
  onopen: Function;
  onmessage: Function;
  send: (val: string) => void;
  onerror: Function;
}

enum GameRoomStatus {
  Lobby,
  WaitingForOpponent,
  Filled
}

enum Rank {
  Ace = 1,
  Two = 2,
  Three = 3,
  Four = 4,
  Five = 5,
  Six = 6,
  Seven = 7,
  Eight = 8,
  Nine = 9,
  Ten = 10,
  Jack = 11,
  Queen = 12,
  King = 13,
}

enum Suit {
  Spades = "spades",
  Clubs = "clubs",
  Hearts = "hearts",
  Diamonds = "diamonds",
}

interface Card {
  suit: Suit;
  rank: Rank;
}


enum GameStatus {
  Starting = "starting",
  BegTurn = "begturn",
  WaitDiscard = "waitdiscard"
}

interface Game {
  turn: number;
  player1: PlayerInfo;
  player2: PlayerInfo;
  deck: Card[];
  player1hand: Card[];
  player2hand: Card[];
  discardpile: Card[];
  status: GameStatus;
  name: string;
}

interface PlayerInfo {
  id: number;
  url: string;
  gameroom: string;
}

interface OutputData {
  messagetype: string;
  command: string;
  content: string;
  playerinfo: PlayerInfo;
  card: Card;
}

interface InputData {
  messagetype: string;
  command: string;
  content: string;
  game: Game
}


const isStartNotification = (data: InputData) => {
  return (
    data["messagetype"] === "meta" &&
    data["command"] === "gameroomstatus" &&
    data["content"] === "filled" &&
    data.game.status === "starting"
  );
};

const isIAmNotification = (data: InputData) => {
  return data["messagetype"] === "meta" && data["command"] === "iam"
}

const isOpponentNotification = (data: InputData) => {
  return data["messagetype"] === "meta" && data["command"] === "opponent"
}

const stringifyCard = (card: Card) => {
  return card.suit + card.rank;
}

const getTurnByID = (game: Game, yourID: number) => {
  return yourID === game.turn
}

const getHandByID = (game: Game, yourID: number) => {
  if (yourID === game.player1.id) {
    return game.player1hand
  }
  if (yourID === game.player2.id) {
    return game.player2hand
  }
  return null
}

// if () {
//   myID = parseInt(data["content"]);
// }

const Game: NextPage = () => {
  //const router = useRouter();
  //const { userInfo, setUserInfo } = useContext(AuthContext);
  const [gameStatus, setGameStatus] = useState<GameRoomStatus>(GameRoomStatus.Lobby)
  const [showWarning, setShowWarning] = useState<boolean>(false)
  const [game, setGame] = useState<Game | null>(null);
  const warning = useRef("")
  const roomName = useRef<string | null>(null);
  const playerName = useRef<string | null>(null);
  const playerID = useRef<number | null>(null);
  const opponentID = useRef<number | null>(null);
  const client = useRef<websocket | null>(null);

  const joinGame = () => {
    if (!roomName.current || roomName.current == "") {
      return;
    }
    client.current = new W3CWebSocket(
      "ws://localhost:8080/channel/" + roomName.current + "/play"
    );

    if (client.current !== null) {
      client.current.onopen = () => {
        setGameStatus(GameRoomStatus.WaitingForOpponent)
      };

      client.current.onmessage = (message: { data: string }) => {
        let data: InputData | null = null;
        try {
          data = JSON.parse(message.data)
        }
        catch (e) {
          console.log("failed to parse non-json message")
          return;
        }

        if (data && GameRoomStatus.Filled === gameStatus) {
          setGame(data.game)
        }

        if (data && isIAmNotification(data)) {
          console.log("Got IAM")
          playerID.current = parseInt(data.content)
        }

        if (data && isOpponentNotification(data)) {
          console.log("Got OPPONENT")
          opponentID.current = parseInt(data.content)
        }

        if (data && GameRoomStatus.Filled === gameStatus) {
          console.log("Game Update!")
          setGame(data.game)
        }

        if (data && isStartNotification(data)) {
          console.log("Got START")
          setGameStatus(GameRoomStatus.Filled)
          setGame(data.game)

        }


        console.log(data)
      };

      client.current.onerror = () => {
        warning.current = "Connection to Server Failed, Try Again Later"
        setShowWarning(true)
      }
    };
  }



  useEffect(() => {
    // if (!userInfo) {
    //   //I SHould really be checking wih the server and not simply if something exists duh
    //   router.push("/");
    // }
  });

  return (
    <div
      className="w-screen h-screen flex justify-center items-center flex-column"
      style={{ backgroundColor: "#FEC5E5", padding:"10px" }}
    >
      {showWarning && <Alert variant='danger'> {warning.current} </Alert>}
      {
        gameStatus && GameRoomStatus.Filled &&
        <>
          <div style={{ display: "flex" }}>
            {game && opponentID.current && getHandByID(game, opponentID.current)?.map((card) => {
              return (
                <div style={{}} key={card.suit + card.rank}>
                  <img width="84" src={"/cards/" + "cardback" + ".png"} alt={stringifyCard(card)} />
                </div>
              )
            })}
          </div>
          <h1
            style={{
              fontWeight: 900,
              fontSize: "3.5rem",
              padding: 30,
              textAlign: "center",
            }}
          >
          {game && playerID.current && getTurnByID(game, playerID.current) && "YOUR TURN"}
          {game && opponentID.current && getTurnByID(game, opponentID.current) && "THEIR TURN"}
          </h1>
          <div style={{ display: "flex", gap: "15px"}}>
            {game && playerID.current && getHandByID(game, playerID.current)?.map((card) => {
              return (
                <div style={{}} key={card.suit + card.rank}>
                  <img width="70" src={"/cards/" + stringifyCard(card) + ".png"} alt={stringifyCard(card)} />
                </div>
              )
            })}
          </div>

        </>


      }
      {gameStatus !== GameRoomStatus.Filled &&
        <>
          <h1
            style={{
              fontWeight: 900,
              fontSize: "3.5rem",
              padding: 30,
              textAlign: "center",
            }}
          >
            GAME LOBBY
          </h1>
        </>
      }
      {gameStatus === GameRoomStatus.Lobby && (
        <Form onSubmit={(e) => e.preventDefault()}>
          <Form.Group className="mb-3" controlId="formBasicEmail">
            <Form.Control
              type="text"
              placeholder="Room Name"
              onChange={(e) => (roomName.current = e.target.value)}
            />
          </Form.Group>
          <Form.Group className="mb-3" controlId="formBasicEmail">
            <Form.Control
              type="text"
              placeholder="User Name"
              onChange={(e) => (playerName.current = e.target.value)}
            />
          </Form.Group>
          <Button
            variant="primary"
            type="submit"
            size="lg"
            onClick={() => joinGame()}
          >
            Join Game
          </Button>
        </Form>
      )}
      {gameStatus === GameRoomStatus.WaitingForOpponent && (
        <h1 style={{ fontWeight: 900, fontSize: "2rem" }}>
          You are currently waiting in {roomName.current}
        </h1>
      )}
    </div>
  );
};


export default Game;
