import { useContext } from "react"
import { UserContext, UserContextType } from "../UserContext"

// Typed accessor for the user context that fails fast when used outside the provider.
export const useUser = (): UserContextType => {
  const context = useContext(UserContext)
  if (!context) {
    throw new Error("useUser must be used within a UserProvider")
  }
  return context
}

