import "../styles/globals.css";
import type { AppProps } from "next/app";
import { useState } from "react";
import AuthContext from '../utils/AuthContext'

function MyApp({ Component, pageProps }: AppProps) {
  const [userInfo, setUserInfo] = useState<String | null>(null);
  return (
    <AuthContext.Provider value={{userInfo, setUserInfo}}>
      <Component {...pageProps} />
    </AuthContext.Provider>
  );
}

export default MyApp;
