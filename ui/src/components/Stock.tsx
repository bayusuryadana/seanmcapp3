import { Fragment } from "react";
import { Grid, IconButton, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Typography } from "@mui/material";
import { Title } from "./Title.tsx";
import { RowActions } from "./RowActions.tsx";
import AddIcon from "@mui/icons-material/Add";
import { WalletStock } from "../utils/model.ts";
import { compactTableStyle, tableContainerStyle } from "../utils/constant.ts";

interface StockProps {
  title: string
  rows: WalletStock[]
  editHandler: (row: WalletStock) => void
  deleteHandler: (row: WalletStock) => void
  createHandler: () => void
  showOwnedColumns?: boolean
}

const getTotalBought = (row: WalletStock) =>
  row.buy_price && row.lot ? row.buy_price * row.lot * 100 : undefined

const getProfitLossPercentage = (row: WalletStock) => {
  if (!row.buy_price || !row.current_price || row.buy_price <= 0) {
    return undefined
  }
  return ((row.current_price - row.buy_price) / row.buy_price) * 100
}

const getProfitLossColor = (value: number | undefined) => {
  if (value === undefined) {
    return 'text.secondary'
  }
  if (value > 0) {
    return 'success.main'
  }
  if (value < 0) {
    return 'error.main'
  }
  return 'warning.main'
}

export const Stock = (props: StockProps) => {
  const showOwnedColumns = props.showOwnedColumns ?? true

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
      <TableContainer sx={tableContainerStyle}>
        <Table size="small" sx={compactTableStyle}>
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>Best Price</TableCell>
              <TableCell>Current Price</TableCell>
              <TableCell>Fair Price</TableCell>
              {showOwnedColumns && <TableCell>Buy Price</TableCell>}
              {showOwnedColumns && <TableCell>Total Bought</TableCell>}
              {showOwnedColumns && <TableCell>P/L</TableCell>}
              <TableCell></TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {props.rows.map((row) => {
              const totalBought = getTotalBought(row)
              const profitLossPercentage = getProfitLossPercentage(row)

              return (
                <TableRow key={row.name}>
                  <TableCell>{row.name}</TableCell>
                  <TableCell>{row.best_price}</TableCell>
                  <TableCell>{row.current_price}</TableCell>
                  <TableCell>{row.fair_price}</TableCell>
                  {showOwnedColumns && <TableCell>{row.buy_price ?? '-'}</TableCell>}
                  {showOwnedColumns && <TableCell>{totalBought !== undefined ? totalBought.toLocaleString() : '-'}</TableCell>}
                  {showOwnedColumns && (
                    <TableCell>
                      <Typography variant="body2" sx={{ color: getProfitLossColor(profitLossPercentage), fontWeight: 600 }}>
                        {profitLossPercentage !== undefined ? `${profitLossPercentage.toFixed(2)}%` : '-'}
                      </Typography>
                    </TableCell>
                  )}
                  <RowActions onEdit={() => props.editHandler(row)} onDelete={() => props.deleteHandler(row)} />
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </TableContainer>
    </Fragment>
  )
}
