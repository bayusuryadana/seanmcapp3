import { Typography } from "@mui/material";
import { ReactNode } from "react";


interface CellTypographyProps {
    done: boolean
    children?: ReactNode;
  }
  
  export const CellTypography = (props: CellTypographyProps) => {

    const getColor = () => {
        if (!props.done) {
            return 'text.primary'
        } else {
            return 'text.secondary'
        }
    }

    return (
        <Typography sx={{color: getColor(), fontSize: { xs: "0.75rem", sm: "0.875rem" }, }} >{props.children}</Typography>
    );
  }
