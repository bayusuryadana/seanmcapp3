import { Typography } from "@mui/material";
import { ReactNode } from "react";

interface CellTypographyProps {
  done: boolean
  children?: ReactNode
}

export const CellTypography = (props: CellTypographyProps) => {
  return (
    <Typography sx={{ color: props.done ? 'text.secondary' : 'text.primary', fontSize: { xs: "0.75rem", sm: "0.875rem" } }}>
      {props.children}
    </Typography>
  );
}
