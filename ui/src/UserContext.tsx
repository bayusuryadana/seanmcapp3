import { ReactNode, useState, FC, createContext, useEffect } from "react";
import { setUnauthorizedHandler } from "./utils/api";

interface Props {
    children: ReactNode
}

export const UserProvider: FC<Props> = ({ children }) => {
    const [userContext, setUserContext] = useState<string | null>(null);

    useEffect(() => {
        const token = localStorage.getItem("token");
        const expiry = localStorage.getItem("tokenExpiry");

        if (token && expiry) {
            const now = Date.now();
            if (now < Number(expiry)) {
                setUserContext(token);
            } else {
                // expired, clear it
                localStorage.removeItem("token");
                localStorage.removeItem("tokenExpiry");
            }
        }
    }, []);

    const saveToken = (token: string | null) => {
        if (token) {
            setUserContext(token);
            localStorage.setItem("token", token);
            localStorage.setItem(
                "tokenExpiry",
                String(Date.now() + 12 * 60 * 60 * 1000) // 12 hours in ms
            );
        } else {
            setUserContext(null);
            localStorage.removeItem("token");
            localStorage.removeItem("tokenExpiry");
        }
    };

    // Let the shared axios instance log the user out on a 401 response.
    useEffect(() => {
        setUnauthorizedHandler(() => saveToken(null));
        return () => setUnauthorizedHandler(null);
    }, []);

    return (
        <UserContext.Provider value={{ userContext, saveToken }}>
            {children}
        </UserContext.Provider>
    );
};

export type UserContextType = {
    userContext: string|null;
    saveToken: (token: string|null) => void;
}

export const UserContext = createContext<UserContextType|null>(null);

