import {Fragment} from "react";
import {Grid, IconButton, Table, TableBody, TableCell, TableContainer, TableHead, TableRow} from "@mui/material";
import {Title} from "./Title.tsx";
import AddIcon from "@mui/icons-material/Add";
import EditIcon from "@mui/icons-material/Edit";
import DeleteIcon from "@mui/icons-material/Delete";
import {WalletStock} from "../utils/model.ts";

interface StockProps {
    title: string
    rows: WalletStock[]
    editHandler: (row: WalletStock) => void
    deleteHandler: (name: string) => void
    createHandler: () => void
}

export const Stock = (props: StockProps) => {
    return (
        <Fragment>
            <Grid container justifyContent={'space-between'}>
                <Grid item>
                    <Title>{props.title}</Title>
                </Grid>
                <Grid item>
                    <IconButton color='primary' size='small' onClick={props.createHandler}>
                        <AddIcon />
                    </IconButton>
                </Grid>
            </Grid>
            <TableContainer sx={{ overflowX: 'auto' }}>
                <Table size="small">
                    <TableHead>
                        <TableRow>
                            <TableCell>Name</TableCell>
                            <TableCell>Best Price</TableCell>
                            <TableCell>Current Price</TableCell>
                            <TableCell>Fair Price</TableCell>
                            <TableCell></TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {props.rows.map((row) => (
                            <TableRow key={row.name}>
                                <TableCell>{row.name}</TableCell>
                                <TableCell>{row.best_price}</TableCell>
                                <TableCell>{row.current_price}</TableCell>
                                <TableCell>{row.fair_price}</TableCell>
                                <TableCell sx={{ whiteSpace: "nowrap" }}>
                                    <IconButton aria-label="edit" color="primary" onClick={()=>props.editHandler(row)}>
                                        <EditIcon />
                                    </IconButton>
                                    <IconButton aria-label="delete" color="secondary" onClick={()=>props.deleteHandler(row.name)}>
                                        <DeleteIcon />
                                    </IconButton>
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </TableContainer>
        </Fragment>
    )
}
