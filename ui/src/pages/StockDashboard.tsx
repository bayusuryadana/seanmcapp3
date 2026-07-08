import { Container, Grid, Paper, Box, Button, CircularProgress, Typography } from "@mui/material"
import RefreshIcon from "@mui/icons-material/Refresh"
import { WalletStock } from "../utils/model.ts"
import { api } from "../utils/api.ts"
import { useEffect, useState } from "react"
import { STOCK_POOL_MONEY, dashboardPaperStyle } from "../utils/constant.ts"
import { Stock } from "../components/Stock.tsx"
import { WalletStockModal } from "../components/WalletStockModal.tsx"
import { AppAlert } from "../components/AppAlert.tsx"
import { useAlert } from "../hooks/useAlert.ts"
import { useModal } from "../hooks/useModal.ts"

export const StockDashboard = () => {

  const { alert, showError, clearAlert } = useAlert()
  const { modal, openCreate, openEdit, openDelete, close } = useModal<WalletStock>()
  const [stocks, setStocks] = useState<WalletStock[]>([])
  const [refreshing, setRefreshing] = useState(false)

  const getStocks = () => {
    api.post('/api/stock/getAll', {})
    .then((response) => {
      clearAlert()
      setStocks(response.data.data ?? [])
    })
    .catch(() => showError('Data failed to fetch/parse!'))
  }

  useEffect(() => {
    getStocks()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const refreshPrices = () => {
    setRefreshing(true)
    api.post('/api/stock/refresh', {})
    .then((response) => {
      clearAlert()
      setStocks(response.data.data ?? [])
    })
    .catch(() => showError('Failed to refresh prices!'))
    .finally(() => setRefreshing(false))
  }

  const onSuccess = () => {
    close()
    getStocks()
  }

  const portfolio = stocks.filter((s) => s.status)
  const wishlist = stocks.filter((s) => !s.status)
  const totalBought = portfolio.reduce((sum, stock) => {
    if (stock.buy_price === undefined || stock.lot === undefined) {
      return sum
    }
    return sum + (stock.buy_price * stock.lot * 100)
  }, 0)
  const remainingMoney = STOCK_POOL_MONEY - totalBought

  return (
    <>
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <AppAlert alert={alert} sx={{ mb: 2 }} />
        <Box sx={{ display: 'flex', justifyContent: 'flex-start', mb: 2 }}>
          <Button
            variant="contained"
            startIcon={refreshing ? <CircularProgress size={18} color="inherit" /> : <RefreshIcon />}
            onClick={refreshPrices}
            disabled={refreshing}
          >
            {refreshing ? 'Refreshing...' : 'Refresh prices'}
          </Button>
        </Box>
        <Grid container spacing={3} sx={{ mb: 3 }}>
          <Grid item md={3} xs={12}>
            <Paper sx={dashboardPaperStyle}>
              <Typography variant="h4">Cash</Typography>
              <Typography color="text.secondary" sx={{ mt: 1 }}>
                Rp {remainingMoney.toLocaleString()}
              </Typography>
            </Paper>
          </Grid>
        </Grid>
        <Grid container spacing={3}>
          <Grid item xs={12} md={7}>
            <Paper sx={dashboardPaperStyle}>
              <Stock
                title="Portfolio"
                rows={portfolio}
                showOwnedColumns={true}
                createHandler={() => openCreate({ name: '', status: true } as WalletStock)}
                editHandler={openEdit}
                deleteHandler={openDelete}
              />
            </Paper>
          </Grid>
          <Grid item xs={12} md={5}>
            <Paper sx={dashboardPaperStyle}>
              <Stock
                title="Wishlist"
                rows={wishlist}
                showOwnedColumns={false}
                createHandler={() => openCreate({ name: '', status: false } as WalletStock)}
                editHandler={openEdit}
                deleteHandler={openDelete}
              />
            </Paper>
          </Grid>
        </Grid>
      </Container>

      <WalletStockModal
        mode={modal?.mode ?? null}
        stock={modal?.item ?? null}
        onClose={close}
        onSuccess={onSuccess}
      />
    </>
  )
}

