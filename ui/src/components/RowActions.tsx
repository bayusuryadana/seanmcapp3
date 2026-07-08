import { IconButton, TableCell } from "@mui/material"
import EditIcon from "@mui/icons-material/Edit"
import DeleteIcon from "@mui/icons-material/Delete"

interface RowActionsProps {
  onEdit: () => void
  onDelete: () => void
}

// Edit/Delete action buttons shared by the wallet and stock tables.
export const RowActions = ({ onEdit, onDelete }: RowActionsProps) => (
  <TableCell sx={{ whiteSpace: "nowrap" }}>
    <IconButton size="small" sx={{ p: 0.25 }} aria-label="edit" color="primary" onClick={onEdit}>
      <EditIcon fontSize="small" />
    </IconButton>
    <IconButton size="small" sx={{ p: 0.25 }} aria-label="delete" color="secondary" onClick={onDelete}>
      <DeleteIcon fontSize="small" />
    </IconButton>
  </TableCell>
)

