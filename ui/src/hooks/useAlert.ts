import { useState } from "react"

export type AlertState = {
  visible: boolean
  text: string
}

const HIDDEN: AlertState = { visible: false, text: "" }

// Small helper to manage a single error/notification message per component.
export const useAlert = () => {
  const [alert, setAlert] = useState<AlertState>(HIDDEN)
  const showError = (text: string) => setAlert({ visible: true, text })
  const clearAlert = () => setAlert(HIDDEN)
  return { alert, showError, clearAlert }
}

