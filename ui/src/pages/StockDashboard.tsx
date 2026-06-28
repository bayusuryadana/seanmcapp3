import { Container, Alert, Grid, Paper, Box, Button, CircularProgress, Typography } from "@mui/material"
import RefreshIcon from "@mui/icons-material/Refresh"
import { WalletStock, WalletAlert } from "../utils/model.ts"
import axios from "axios"
import { useContext, useEffect, useState } from "react"
import { UserContext, UserContextType } from "../UserContext.tsx"
import { API_URL, STOCK_POOL_MONEY } from "../utils/constant.ts"
import { Stock } from "../components/Stock.tsx"
import { WalletStockModal } from "../components/WalletStockModal.tsx"

export const StockDashboard = () => {

  const { userContext, saveToken } = useContext(UserContext) as UserContextType
  const [alert, setAlert] = useState<WalletAlert>({ display: 'none', text: '' })
  const [stocks, setStocks] = useState<WalletStock[]>([])
  const [walletStock, setWalletStock] = useState<WalletStock | null>(null)
  const [refreshing, setRefreshing] = useState(false)

  const getStocks = () => {
    axios.post(API_URL + '/api/stock/getAll', {}, {
      headers: {
        Authorization: 'Bearer ' + (userContext ?? "")
      },
    })
    .then((response) => {
      setAlert({ display: 'none', text: '' })
      setStocks(response.data.data ?? [])
    })
    .catch((error) => {
      console.log(error)
      if (axios.isAxiosError(error) && error.response?.status == 401) {
        saveToken(null)
      } else {
        setAlert({ display: 'true', text: 'Data failed to fetch/parse!' })
      }
    })
  }

  useEffect(() => {
    getStocks()
  }, [])

  const refreshPrices = () => {
    setRefreshing(true)
    axios.post(API_URL + '/api/stock/refresh', {}, {
      headers: {
        Authorization: 'Bearer ' + (userContext ?? "")
      },
    })
    .then((response) => {
      setAlert({ display: 'none', text: '' })
      setStocks(response.data.data ?? [])
    })
    .catch((error) => {
      console.log(error)
      if (axios.isAxiosError(error) && error.response?.status == 401) {
        saveToken(null)
      } else {
        setAlert({ display: 'true', text: 'Failed to refresh prices!' })
      }
    })
    .finally(() => setRefreshing(false))
  }

  const onSuccess = () => {
    setWalletStock(null)
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
        <Alert id="invalid-data-alert" severity="error" sx={{ mb: 2, display: alert.display }}>{alert.text}</Alert>
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
            <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
              <Typography variant="h4">Cash</Typography>
              <Typography color="text.secondary" sx={{ mt: 1 }}>
                Rp {remainingMoney.toLocaleString()}
              </Typography>
            </Paper>
          </Grid>
        </Grid>
        <Grid container spacing={3}>
          <Grid item xs={12} md={7}>
            <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
              <Stock
                title="Portfolio"
                rows={portfolio}
                showOwnedColumns={true}
                createHandler={() => setWalletStock({ name: '', status: true } as WalletStock)}
                editHandler={(stock: WalletStock) => setWalletStock(stock)}
                deleteHandler={(name: string) => setWalletStock({ name: name } as WalletStock)}
              />
            </Paper>
          </Grid>
          <Grid item xs={12} md={5}>
            <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
              <Stock
                title="Wishlist"
                rows={wishlist}
                showOwnedColumns={false}
                createHandler={() => setWalletStock({ name: '', status: false } as WalletStock)}
                editHandler={(stock: WalletStock) => setWalletStock(stock)}
                deleteHandler={(name: string) => setWalletStock({ name: name } as WalletStock)}
              />
            </Paper>
          </Grid>
        </Grid>
      </Container>

      <WalletStockModal
        onClose={() => setWalletStock(null)}
        onSuccess={onSuccess}
        walletStock={walletStock}
      />
    </>
  )
}

