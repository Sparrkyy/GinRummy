import type { NextPage } from "next";
import AuthContext from "./AuthContext";
import { useRouter } from "next/router";
import { useContext, useEffect, useState } from "react";
import Button from "react-bootstrap/button";
import "bootstrap/dist/css/bootstrap.min.css";
import { w3cwebsocket as W3CWebSocket } from "websocket";

const client = new W3CWebSocket("ws://localhost:8080/play")

const Game: NextPage = () => {
  const router = useRouter();
  const { userInfo, setUserInfo } = useContext(AuthContext);
  const [messages, setMessages] = useState<string[]>([]);

  useEffect(() => {
    const myID = -1;
    client.onopen = () => {
      console.log("opened") 
    }
    client.onmessage = (message:string) => {
      console.log(message) 
    }

  });

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
      <Button variant="primary" type="submit" size="lg">
        Join Game
      </Button>
      <div>
        {messages.map((ele) => (
          <p>{ele}</p>
        ))}
      </div>
    </div>
  );
};

export default Game;
