import { ReactNode, useState, FC, createContext } from "react";

interface Props {
    children: ReactNode
}

export const UserProvider: FC<Props> = ({ children }) => {
    const [userContext, setToken] = useState<string|null>(() => {
        return localStorage.getItem('token');
    })

    const saveToken = (token: string|null) => {
        setToken(token);
        if (token) {
            localStorage.setItem('token', token);
        } else {
            localStorage.removeItem('token');
        }
    }

    return <UserContext.Provider value={{ userContext, saveToken}}>
        {children}
    </UserContext.Provider>
}

export type UserContextType = {
    userContext: string|null;
    saveToken: (token: string|null) => void;
}

export const UserContext = createContext<UserContextType|null>(null);

