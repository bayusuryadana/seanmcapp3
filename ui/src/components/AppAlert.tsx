import { Alert, AlertColor, SxProps, Theme } from "@mui/material"
import { AlertState } from "../hooks/useAlert"

interface AppAlertProps {
  alert: AlertState
  severity?: AlertColor
  sx?: SxProps<Theme>
}

// Renders an alert only when there is a message to show.
export const AppAlert = ({ alert, severity = "error", sx }: AppAlertProps) => {
  if (!alert.visible) {
    return null
  }
  return (
    <Alert severity={severity} sx={sx}>
      {alert.text}
    </Alert>
  )
}

