import type { NextPage } from "next";
import AuthContext from "./AuthContext";
import { useRouter } from "next/router";
import { useContext, useEffect } from "react";
import Button from "react-bootstrap/button"
import "bootstrap/dist/css/bootstrap.min.css";

const Game: NextPage = () => {
  const router = useRouter();
  const { userInfo, setUserInfo } = useContext(AuthContext);

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
    </div>
  );
};

export default Game;
