import "../styles/globals.css";
import type { AppProps } from "next/app";
import { useState } from "react";
import { createContext } from "react";
import type { SetStateAction, Dispatch} from "react"

const AuthContext = createContext<{ userInfo: String | null; setUserInfo: Dispatch<SetStateAction<String | null>>; }|null>(null);

function MyApp({ Component, pageProps }: AppProps) {
  const [userInfo, setUserInfo] = useState<String | null>(null);
  return (
    <AuthContext.Provider value={{userInfo, setUserInfo}}>
      <Component {...pageProps} />
    </AuthContext.Provider>
  );
}

export default MyApp;
