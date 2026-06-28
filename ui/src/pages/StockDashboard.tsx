import { Container, Alert, Grid, Paper } from "@mui/material"
import { WalletStock, WalletAlert } from "../utils/model.ts"
import axios from "axios"
import { useContext, useEffect, useState } from "react"
import { UserContext, UserContextType } from "../UserContext.tsx"
import { API_URL } from "../utils/constant.ts"
import { Stock } from "../components/Stock.tsx"
import { WalletStockModal } from "../components/WalletStockModal.tsx"

export const StockDashboard = () => {

  const { userContext, saveToken } = useContext(UserContext) as UserContextType
  const [alert, setAlert] = useState<WalletAlert>({ display: 'none', text: '' })
  const [stocks, setStocks] = useState<WalletStock[]>([])
  const [walletStock, setWalletStock] = useState<WalletStock | null>(null)

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

  const onSuccess = (row: WalletStock, actionText: string | undefined) => {
    setWalletStock(null)
    if (actionText === 'Create') {
      setStocks([...stocks, row])
    } else if (actionText === 'Edit') {
      // keep the existing current_price (auto-fetched by the backend scheduler)
      setStocks(stocks.map((s) => s.name === row.name ? { ...s, ...row } : s))
    } else if (actionText === 'Delete') {
      setStocks(stocks.filter((s) => s.name !== row.name))
    }
  }

  return (
    <>
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Alert id="invalid-data-alert" severity="error" sx={{ mb: 2, display: alert.display }}>{alert.text}</Alert>
        <Grid container spacing={3}>
          <Grid item xs={12}>
            <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
              <Stock
                rows={stocks}
                createHandler={() => setWalletStock({ name: '' } as WalletStock)}
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

