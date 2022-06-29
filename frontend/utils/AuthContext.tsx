import type { SetStateAction, Dispatch} from "react"
import { createContext } from "react";

const AuthContext = createContext<{ userInfo: String | null; setUserInfo: Dispatch<SetStateAction<String | null>>; }|null>(null);

export default AuthContext
