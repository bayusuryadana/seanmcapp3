import { Fragment, useState } from "react";
import { Button, Grid, IconButton, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TableSortLabel, Typography } from "@mui/material";
import { Title } from "./Title.tsx";
import { RowActions } from "./RowActions.tsx";
import AddIcon from "@mui/icons-material/Add";
import { DashboardPerformance, WalletStock } from "../utils/model.ts";
import { compactTableStyle, tableContainerStyle } from "../utils/constant.ts";

interface StockProps {
  title: string
  rows: WalletStock[]
  editHandler: (row: WalletStock) => void
  deleteHandler: (row: WalletStock) => void
  createHandler: () => void
  showOwnedColumns?: boolean
  showTitle?: boolean
  performance?: Record<string, DashboardPerformance>
  period?: string
  onPeriodChange?: (period: string) => void
  maskMoney?: boolean
}

const getTotalBought = (row: WalletStock) =>
  row.buy_price !== undefined && row.lot !== undefined ? row.buy_price * row.lot * 100 : undefined

const getProfitLossPercentage = (row: WalletStock) => {
  if (row.buy_price === undefined || row.current_price === undefined || row.buy_price <= 0) {
    return undefined
  }
  return ((row.current_price - row.buy_price) / row.buy_price) * 100
}

const getProfitLossAmount = (row: WalletStock) => {
  if (row.buy_price === undefined || row.current_price === undefined || row.lot === undefined) {
    return undefined
  }
  return (row.current_price - row.buy_price) * row.lot * 100
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
  const showTitle = props.showTitle ?? true
  const [orderBy, setOrderBy] = useState<string>('name')
  const [order, setOrder] = useState<'asc' | 'desc'>('asc')

  const changeSort = (key: string) => {
    setOrder(orderBy === key && order === 'asc' ? 'desc' : 'asc')
    setOrderBy(key)
  }
  const valueFor = (row: WalletStock, key: string) => {
    if (key === 'totalBought') return getTotalBought(row)
    if (key === 'profitLossPercentage') return props.performance?.[row.name]?.percentage ?? getProfitLossPercentage(row)
    if (key === 'profitLossAmount') return props.performance?.[row.name]?.delta ?? getProfitLossAmount(row)
    return row[key as keyof WalletStock]
  }
  const rows = [...props.rows].sort((a, b) => {
    const aValue = valueFor(a, orderBy)
    const bValue = valueFor(b, orderBy)
    const compared = typeof aValue === 'string' || typeof bValue === 'string'
      ? String(aValue ?? '').localeCompare(String(bValue ?? ''))
      : Number(aValue ?? -Infinity) - Number(bValue ?? -Infinity)
    return order === 'asc' ? compared : -compared
  })
  const header = (label: string, key: string, sx = {}) => (
    <TableCell sortDirection={orderBy === key ? order : false} sx={sx}>
      <TableSortLabel active={orderBy === key} direction={orderBy === key ? order : 'asc'} onClick={() => changeSort(key)}>{label}</TableSortLabel>
    </TableCell>
  )

  return (
    <Fragment>
      <Grid container justifyContent={'space-between'}>
        {showTitle && <Grid item><Title>{props.title}</Title></Grid>}
        {props.onPeriodChange && (
          <Grid item sx={{ display: 'flex', alignItems: 'center' }}>
            {['all', '1d', '5d', '1mo', '3mo', '6mo', '1y', 'ytd'].map((period) => (
            <Button key={period} size="small" variant={props.period === period ? 'contained' : 'text'} onClick={() => props.onPeriodChange?.(period)}>
              {period === 'all' ? 'Overall' : period.toUpperCase()}
            </Button>
            ))}
          </Grid>
        )}
      </Grid>
      <TableContainer sx={tableContainerStyle}>
        <Table size="small" sx={compactTableStyle}>
          <TableHead>
            <TableRow>
              {header('Name', 'name')}
              {header('Best Price', 'best_price')}
              {header('Current Price', 'current_price')}
              {header('Fair Price', 'fair_price')}
              {showOwnedColumns && header('Buy Price', 'buy_price')}
              {showOwnedColumns && header('Total Bought', 'totalBought')}
              {showOwnedColumns && (
                <>
                  {header('P/L %', 'profitLossPercentage')}
                  {header('P/L Rp', 'profitLossAmount', { minWidth: 110, whiteSpace: 'nowrap' })}
                </>
              )}
              <TableCell align="right"><IconButton color="primary" size="small" aria-label="add stock" onClick={props.createHandler}><AddIcon /></IconButton></TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {rows.map((row) => {
              const totalBought = getTotalBought(row)
              const periodPerformance = props.performance?.[row.name]
              const profitLossPercentage = periodPerformance?.percentage ?? getProfitLossPercentage(row)
              const profitLossAmount = periodPerformance?.delta ?? getProfitLossAmount(row)

              return (
                <TableRow key={row.name}>
                  <TableCell>{row.name}</TableCell>
                  <TableCell>{row.best_price}</TableCell>
                  <TableCell>{row.current_price}</TableCell>
                  <TableCell>{row.fair_price}</TableCell>
                  {showOwnedColumns && <TableCell>{row.buy_price ?? '-'}</TableCell>}
                  {showOwnedColumns && <TableCell>{totalBought !== undefined ? (props.maskMoney ? '••••••' : totalBought.toLocaleString()) : '-'}</TableCell>}
                  {showOwnedColumns && (
                    <>
                      <TableCell><Typography variant="body2" sx={{ color: getProfitLossColor(profitLossPercentage), fontWeight: 600 }}>{profitLossPercentage === undefined ? '-' : `${profitLossPercentage.toFixed(2)}%`}</Typography></TableCell>
                      <TableCell sx={{ minWidth: 110, whiteSpace: 'nowrap' }}><Typography variant="body2" sx={{ color: getProfitLossColor(profitLossAmount), fontWeight: 600 }}>{profitLossAmount === undefined ? '-' : props.maskMoney ? 'Rp ••••••' : `Rp ${profitLossAmount.toLocaleString()}`}</Typography></TableCell>
                    </>
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
