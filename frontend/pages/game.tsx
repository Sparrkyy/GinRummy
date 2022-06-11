import type { NextPage } from "next";
import AuthContext from "./AuthContext";
import { useRouter } from "next/router";
import { useContext, useEffect, useState, useRef } from "react";
import Button from "react-bootstrap/button";
import Form from "react-bootstrap/Form";
import "bootstrap/dist/css/bootstrap.min.css";
import { w3cwebsocket as W3CWebSocket } from "websocket";

interface websocket {
  onopen: Function;
  onmessage: Function;
  send: (val: string) => void;
}

const Game: NextPage = () => {
  const router = useRouter();
  const { userInfo, setUserInfo } = useContext(AuthContext);
  const [messages, setMessages] = useState<string[]>([]);
  const [inGame, setInGame] = useState<boolean>(false);
  const roomName = useRef<string | null>(null);
  const client = useRef<websocket|null>(null);

  const joinGame = () => {
    if (!roomName.current || roomName.current == "") {
      return;
    }
    client.current = new W3CWebSocket(
      "ws://localhost:8080/channel/" + roomName.current + "/play"
    );
    if (client.current !== null) {
      client.current.onopen = () => {
        setInGame(true);
      };

      client.current.onmessage = (message: { data: string }) => {
        setMessages((prev) => [...prev, message.data]);
      };
    }
  };

  useEffect(() => {
    if (!userInfo) {
      //I SHould really be checking wih the server and not simply if something exists duh
      //router.push("/");
    }
  });

  return (
    <div
      className="w-screen h-screen flex justify-center items-center flex-column"
      style={{ backgroundColor: "#FEC5E5" }}
    >
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
      {!inGame && (
        <Form onSubmit={(e) => e.preventDefault()}>
          <Form.Group className="mb-3" controlId="formBasicEmail">
            <Form.Control
              type="text"
              placeholder="Room Name"
              onChange={(e) => (roomName.current = e.target.value)}
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
      {inGame && (
        <h1 style={{ fontWeight: 900, fontSize: "2rem" }}>
          {roomName.current}
        </h1>
      )}
      {inGame && (
        <button
          onClick={() => {
            console.log("clicked")
            if (client.current !== null) {
              console.log("sent")
              client.current.send("hello");
            }
          }}
        >
          Send Hello
        </button>
      )}
      <div>
        {messages.map((ele) => (
          <p key={ele}>{ele}</p>
        ))}
      </div>
    </div>
  );
};

export default Game;
