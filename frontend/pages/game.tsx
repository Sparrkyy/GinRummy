import type { NextPage } from "next";
import AuthContext from "./AuthContext";
import { useRouter } from "next/router";
import { useContext, useEffect, useState, useRef } from "react";
import Button from "react-bootstrap/button";
import Form from "react-bootstrap/Form";
import Alert from "react-bootstrap/Alert"
import "bootstrap/dist/css/bootstrap.min.css";
import { w3cwebsocket as W3CWebSocket } from "websocket";
import { DragDropContext, Droppable, Draggable } from 'react-beautiful-dnd';


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
  card?: Card;
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

const isGameUpdateNotification = (data: InputData) => {
  return data.command === "gameupdate"
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

const getFirstDiscard = (game: Game) => {
  const discardPile = game.discardpile;
  return discardPile[discardPile.length - 1]
}

const areCardsEqual = (card1: Card, card2: Card | null) =>{
  if (!card2) return false
  if (card1.rank !== card2.rank) return false
  if (card1.suit !== card2.suit) return false
  return true
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
  const [selectedCard, setSelectedCard] = useState<Card | null>(null);
  const [hand, setHand] = useState<Card[] | null>(null);
  const dragCard = useRef<null | number>(null);
  const dragOverCard = useRef<null | number>(null);



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

        if (data && isGameUpdateNotification(data)) {
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
        setFadedWarning("Connection to Server Failed, Try Again Later")
      }
    };
  }

  const drawCard = (playerinfo: PlayerInfo, option: string) => {
    const response: OutputData = { messagetype: "game", command: "draw", content: option, playerinfo: playerinfo }
    if (client.current) {
      client.current.send(JSON.stringify(response))
    }
    else {
      setFadedWarning("No Connection: Try Again Later")
    }
  }

  const discardSelectedCard = () => {
    if (client.current && game && selectedCard && playerID) {
      const response: OutputData = {
        messagetype: "game",
        command: "discard",
        content: "",
        card: selectedCard,
        playerinfo: playerID.current === game.player1.id ? game.player1 : game.player2,
      };
      client.current.send(JSON.stringify(response))
    }
    else {
      setFadedWarning("Error: no connection, or no selection was made")
    }


  }

  function handleOnDragEnd(result: any) {
    if (!result.destination) return;
    if (!hand) return;
    const items = [...hand];
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);
    setHand(items);
  }

  const dragStart = (e:any, position:number) => {
    dragCard.current = position;
  };
  
  const dragEnter = (e:any, position:number) => {
    dragOverCard.current = position;
  };

  const drop = (e:any) => {
    console.log(dragOverCard.current, dragCard.current)
    if (!hand) return;
    if (dragCard.current === null) return;
    if (dragOverCard.current === null) return;
    const handCopy = [...hand];
    const dragCardContent = handCopy[dragCard.current];
    handCopy.splice(dragCard.current, 1);
    handCopy.splice(dragOverCard.current, 0, dragCardContent);
    dragCard.current = null;
    dragOverCard.current = null;
    setHand(handCopy);
  };

  const setHandPreserveOrder = (serverHand:Card[]) =>{
    if (!hand){
      setHand(serverHand)
      return
    }
    const cardMap = new Map<string, number>()
    const handCopy = [...hand]
    for (let i=0; i<handCopy.length;i++){
      cardMap.set(stringifyCard(handCopy[i]), i+1)
    }

    const resultHand = [...serverHand]
    resultHand.sort((a:Card, b:Card)=>{
      const scoreA = cardMap.get(stringifyCard(a))
      const scoreB = cardMap.get(stringifyCard(b))
      if (scoreA && scoreB){
        return scoreA - scoreB
      }
      else if (scoreA && !scoreB){
        return 1
      }
      else if (scoreB && !scoreA){
        return 1
      }
      return 0
    })
    setHand(resultHand)
  }

  useEffect(() => {
    console.log("fire use effect")
    if(!game) return;
    console.log("game is not null")
    if (playerID.current === game.player1.id) {
      console.log("is player 1")
      setHandPreserveOrder(game.player1hand)
      //setHand(game.player1hand)
    }
    else if (playerID.current === game.player2.id) {
      console.log("is player 2", game.player2hand)
      setHandPreserveOrder(game.player2hand)
      // setHand(game.player2hand)
    }
    else {
      console.log("is niether current players")
    }
  }, [game])

  const setFadedWarning = (message: string) => {
    warning.current = message
    setShowWarning(true)
    setTimeout(()=>{
      warning.current = ""
      setShowWarning(false)
    }, 2000)
  }

  const toggleSelectedCard = (card: Card) => {
    if (areCardsEqual(card, selectedCard)){
      setSelectedCard(null)
      return
    }
    setSelectedCard(card)
  }



  useEffect(() => {
    // if (!userInfo) {
    //   //I SHould really be checking wih the server and not simply if something exists duh
    //   router.push("/");
    // }
  });

  return (
    <div className="w-screen h-screen flex justify-center items-center flex-column" style={{ backgroundColor: "#FEC5E5", padding: "10px" }}>
      {showWarning && <Alert variant='danger'> {warning.current} </Alert>}
      {gameStatus && GameRoomStatus.Filled &&
        <div>
          <div style={{ display: "flex", flexDirection: "column", alignItems: "center" }}>
            <div style={{ display: "flex" }}>
              {game && opponentID.current && getHandByID(game, opponentID.current)?.map((card) => {
                return (
                  <div style={{}} key={card.suit + card.rank}>
                    <img width="84" src={"/cards/" + "cardback" + ".png"} alt={"card back"} />
                  </div>
                )
              })}
            </div>
            <div style={{ display: "flex", alignItems: "center" }}>
              {game && <img style={{ height: "110px" }} src={"/cards/" + "cardback" + ".png"} alt={"card back"} />}
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
              {game && game.discardpile.length > 0 && <img style={{ height: "100px" }} src={"/cards/" + stringifyCard(getFirstDiscard(game)) + ".png"} alt={stringifyCard(getFirstDiscard(game))} />}
              {game && game.discardpile.length === 0 && <div style={{ height: "100px", border: "2px solid black", textAlign: "center", borderRadius: "5px", padding: "10px" }}>Empty Discard</div>}
            </div>
            <div style={{ display: "flex", gap: "15px" }}>
              {game && playerID.current && hand &&
                hand.map((card, i) => {
                  return (
                    <div style={{cursor:"pointer"}} className={ areCardsEqual(card, selectedCard)? "selected-card": ""} key={i} 
                    onClick={() => toggleSelectedCard(card)}
                    onDragStart={(e)=>dragStart(e,i)}
                    onDragEnter={(e)=>dragEnter(e,i)}
                    onDragOver={(e)=>e.preventDefault()}
                    onDragEnd={drop}
                    draggable
                    > 
                    <img width="70" src={"/cards/" + stringifyCard(card) + ".png"} alt={stringifyCard(card)} /> 
                    </div>
                  )

                })
              }
            </div>
          </div>
          {game && playerID.current && getTurnByID(game, playerID.current) &&
            <div style={{ display: "flex", padding: "15px", justifyContent: "space-around" }}>
              {game.status === "waitdiscard" &&
                <Button
                  variant="primary"
                  type="submit"
                  size="lg"
                  onClick={() => discardSelectedCard()}
                >
                  Discard Selected Card
                </Button>
              }
              {(game.status === "starting" || game.status === "begturn") &&
                <>
                  <Button
                    variant="primary"
                    type="submit"
                    size="lg"
                    onClick={() => drawCard(playerID.current === game.player1.id ? game.player1 : game.player2, "stack")}
                  >
                    Draw Deck
                  </Button>
                  <Button
                    variant="primary"
                    type="submit"
                    size="lg"
                    onClick={() => drawCard(playerID.current === game.player1.id ? game.player1 : game.player2, "discard")}
                  >
                    Draw Discard
                  </Button>
                </>
              }
            </div>

          }
        </div>
      }
      {gameStatus !== GameRoomStatus.Filled && <h1 style={{ fontWeight: 900, fontSize: "3.5rem", padding: 30, textAlign: "center" }}>
        GAME LOBBY
      </h1>
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
