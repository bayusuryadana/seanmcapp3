import { Box, Button, Modal, Typography } from "@mui/material"
import { FormEvent, ReactNode } from "react"
import { modalStyle } from "../utils/constant"

interface FormModalProps {
  open: boolean
  title: string
  submitLabel: string
  onClose: () => void
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
  children: ReactNode
}

// Shared shell for the create/edit/delete dialogs: modal box, title, form and submit button.
export const FormModal = ({ open, title, submitLabel, onClose, onSubmit, children }: FormModalProps) => (
  <Modal open={open} onClose={onClose} aria-label={title}>
    <Box sx={modalStyle}>
      <Typography variant="h6" component="h2">
        {title}
      </Typography>
      <Box component="form" onSubmit={onSubmit} sx={{ mt: 2 }}>
        {children}
        <Button type="submit" fullWidth variant="contained" sx={{ mt: 3, mb: 2 }}>
          {submitLabel}
        </Button>
      </Box>
    </Box>
  </Modal>
)

