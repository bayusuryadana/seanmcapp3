import { Accordion, AccordionDetails, AccordionSummary, Container, Grid, Box, Button, CircularProgress, IconButton, Stack, Typography } from "@mui/material"
import RefreshIcon from "@mui/icons-material/Refresh"
import VisibilityIcon from "@mui/icons-material/Visibility"
import VisibilityOffIcon from "@mui/icons-material/VisibilityOff"
import ExpandMoreIcon from "@mui/icons-material/ExpandMore"
import { StockProgressPoint, StockSummary, WalletStock } from "../utils/model.ts"
import { api } from "../utils/api.ts"
import { useEffect, useState } from "react"
import { STOCK_POOL_MONEY, dashboardPaperStyle } from "../utils/constant.ts"
import { Stock } from "../components/Stock.tsx"
import { WalletStockModal } from "../components/WalletStockModal.tsx"
import { AppAlert } from "../components/AppAlert.tsx"
import { useAlert } from "../hooks/useAlert.ts"
import { useModal } from "../hooks/useModal.ts"
import { CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts"

export const StockDashboard = () => {

  const { alert, showError, clearAlert } = useAlert()
  const { modal, openCreate, openEdit, openDelete, close } = useModal<WalletStock>()
  const [stocks, setStocks] = useState<WalletStock[]>([])
  const [summary, setSummary] = useState<StockSummary | null>(null)
  const [period, setPeriod] = useState('1d')
  const [tableSummary, setTableSummary] = useState<StockSummary | null>(null)
  const [tablePeriod, setTablePeriod] = useState('all')
  const [progression, setProgression] = useState<StockProgressPoint[]>([])
  const [refreshing, setRefreshing] = useState(false)
  const [moneyVisible, setMoneyVisible] = useState(true)

  const getStocks = () => {
    api.post('/api/stock/getAll', {})
    .then((response) => {
      clearAlert()
      setStocks(response.data.data ?? [])
    })
    .catch(() => showError('Data failed to fetch/parse!'))

    getSummary(period)
    getProgression(tablePeriod)
  }

  const getSummary = (selectedPeriod: string) => {
    api.post(`/api/stock/summary?period=${selectedPeriod}`, {})
    .then((response) => setSummary(response.data.data ?? null))
    .catch(() => setSummary(null))
  }

  const changePeriod = (selectedPeriod: string) => {
    setPeriod(selectedPeriod)
    getSummary(selectedPeriod)
  }

  const getProgression = (selectedPeriod: string) => {
    api.post(`/api/stock/progression?period=${selectedPeriod}`, {})
    .then((response) => setProgression(response.data.data ?? []))
    .catch(() => setProgression([]))
  }

  const changeTablePeriod = (selectedPeriod: string) => {
    setTablePeriod(selectedPeriod)
    getProgression(selectedPeriod)
    if (selectedPeriod === 'all') {
      setTableSummary(null)
      changePeriod('1d')
      return
    }
    api.post(`/api/stock/summary?period=${selectedPeriod}`, {})
    .then((response) => setTableSummary(response.data.data ?? null))
    .catch(() => setTableSummary(null))
    changePeriod(selectedPeriod)
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
      getSummary(period)
      getProgression(tablePeriod)
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
  const jkseDelta = summary?.jkse?.delta
  const jkseDeltaPercentage = summary?.jkse?.percentage
  const portfolioDelta = summary?.portfolio?.delta
  const portfolioDeltaPercentage = summary?.portfolio?.percentage

  const signedNumber = (value: number) => `${value > 0 ? '+' : ''}${value.toLocaleString()}`
  const signedPercentage = (value: number) => `${value > 0 ? '+' : ''}${value.toFixed(2)}%`
  const changeColor = (value: number | undefined) => {
    if (value === undefined || value === 0) return 'text.secondary'
    return value > 0 ? 'success.main' : 'error.main'
  }

  return (
    <>
      <Container maxWidth="lg" sx={{ mt: { xs: 2, sm: 4 }, mb: { xs: 2, sm: 4 }, px: { xs: 2, sm: 3 } }}>
        <AppAlert alert={alert} sx={{ mb: 2 }} />
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 1, mb: 2 }}>
          <Box sx={{ display: 'flex', overflowX: 'auto', whiteSpace: 'nowrap', WebkitOverflowScrolling: 'touch', flexGrow: 1 }}>
            {['all', '1d', '5d', '1mo', '3mo', '6mo', '1y', 'ytd'].map((item) => <Button key={item} size="small" variant={tablePeriod === item ? 'contained' : 'text'} onClick={() => changeTablePeriod(item)}>{item === 'all' ? 'Overall' : item.toUpperCase()}</Button>)}
          </Box>
          <Stack direction="row" spacing={0.5} alignItems="center">
            <IconButton aria-label={moneyVisible ? 'Hide money values' : 'Show money values'} onClick={() => setMoneyVisible((visible) => !visible)}>
              {moneyVisible ? <VisibilityOffIcon /> : <VisibilityIcon />}
            </IconButton>
            <Button
              size="small"
              variant="contained"
              startIcon={refreshing ? <CircularProgress size={18} color="inherit" /> : <RefreshIcon />}
              onClick={refreshPrices}
              disabled={refreshing}
            >
              <Box component="span" sx={{ display: { xs: 'none', sm: 'inline' } }}>{refreshing ? 'Refreshing...' : 'Refresh prices'}</Box>
            </Button>
          </Stack>
        </Box>
        <Grid container spacing={3}>
          <Grid item md={12} xs={12}>
            <Accordion defaultExpanded sx={{ ...dashboardPaperStyle, p: 0 }}>
              <AccordionSummary expandIcon={<ExpandMoreIcon />} aria-controls="summary-content" id="summary-header" sx={{ px: 2 }}>
                <Typography variant="h5" sx={{ '&::after': { content: '""', display: 'block', width: 30, height: 2, mt: 0.75, borderRadius: 2, bgcolor: 'primary.main', opacity: 0.55 } }}>Summary</Typography>
              </AccordionSummary>
              <AccordionDetails id="summary-content" sx={{ px: 2, pt: 0, pb: 2 }}>
                <Stack direction={{ xs: 'column', md: 'row' }} spacing={3} alignItems="stretch">
                  <Stack spacing={1.25} sx={{ width: { md: 230 }, flexShrink: 0 }}>
                    <Box>
                      <Typography variant="caption" color="text.secondary">Cash</Typography>
                      <Typography variant="h6">{moneyVisible ? `Rp ${remainingMoney.toLocaleString()}` : 'Rp ••••••'}</Typography>
                    </Box>
                    <Box>
                      <Typography variant="caption" color="text.secondary">Index</Typography>
                      <Typography variant="h6" sx={{ color: changeColor(jkseDelta), fontWeight: 600 }}>
                        {jkseDelta !== undefined && jkseDeltaPercentage !== undefined ? (moneyVisible ? `${signedNumber(jkseDelta)} (${signedPercentage(jkseDeltaPercentage)})` : '••••••') : '-'}
                      </Typography>
                    </Box>
                    <Box>
                      <Typography variant="caption" color="text.secondary">Porto</Typography>
                      <Typography variant="h6" sx={{ color: changeColor(portfolioDelta), fontWeight: 600 }}>
                        {portfolioDelta !== undefined && portfolioDeltaPercentage !== undefined ? (moneyVisible ? `Rp ${signedNumber(portfolioDelta)} (${signedPercentage(portfolioDeltaPercentage)})` : 'Rp ••••••') : '-'}
                      </Typography>
                    </Box>
                  </Stack>
                  <Box sx={{ flexGrow: 1, minWidth: 0, borderLeft: { md: 1 }, borderTop: { xs: 1, md: 0 }, borderColor: 'divider', pl: { md: 3 }, pt: { xs: 2, md: 0 } }}>
                <Typography variant="subtitle2" color="text.secondary" sx={{ mb: 1 }}>Progression</Typography>
                {progression.length > 0 ? (
                  <ResponsiveContainer width="100%" height={220}>
                    <LineChart data={progression} margin={{ top: 8, right: 12, left: 0, bottom: 4 }}>
                      <CartesianGrid horizontal={false} vertical stroke="rgba(128,128,128,0.16)" />
                      <XAxis dataKey="date" axisLine={{ stroke: 'rgba(128,128,128,0.28)' }} tickLine={{ stroke: 'rgba(128,128,128,0.22)' }} tickFormatter={(date) => new Date(`${date}T00:00:00`).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })} minTickGap={32} />
                      <YAxis tickFormatter={(value) => `${value.toFixed(0)}%`} width={46} />
                      <Tooltip labelFormatter={(date) => new Date(`${date}T00:00:00`).toLocaleDateString()} formatter={(value: number) => `${value.toFixed(2)}%`} />
                      <Legend />
                      <Line type="monotone" dataKey="index" name="Index" stroke="rgba(144, 164, 174, 0.8)" strokeWidth={2} strokeDasharray="6 5" dot={false} />
                      <Line type="monotone" dataKey="portfolio" name="Portfolio" stroke="#66bb6a" strokeWidth={2} dot={false} />
                    </LineChart>
                  </ResponsiveContainer>
                ) : (
                  <Typography color="text.secondary" variant="body2">Price progression will appear after prices are available.</Typography>
                )}
              </Box>
                </Stack>
              </AccordionDetails>
            </Accordion>
          </Grid>
          <Grid item xs={12} md={12}>
            <Stack spacing={2}>
              <Accordion defaultExpanded>
                <AccordionSummary expandIcon={<ExpandMoreIcon />} aria-controls="portfolio-content" id="portfolio-header">
                  <Typography variant="h5" sx={{ '&::after': { content: '""', display: 'block', width: 30, height: 2, mt: 0.75, borderRadius: 2, bgcolor: 'primary.main', opacity: 0.55 } }}>Portfolio</Typography>
                </AccordionSummary>
                <AccordionDetails id="portfolio-content">
                  <Stock
                    title="Portfolio"
                    showTitle={false}
                    rows={portfolio}
                    showOwnedColumns={true}
                    performance={tablePeriod === 'all' ? undefined : tableSummary?.positions}
                    maskMoney={!moneyVisible}
                    createHandler={() => openCreate({ name: '', status: true } as WalletStock)}
                    editHandler={openEdit}
                    deleteHandler={openDelete}
                  />
                </AccordionDetails>
              </Accordion>
              <Accordion defaultExpanded>
                <AccordionSummary expandIcon={<ExpandMoreIcon />} aria-controls="wishlist-content" id="wishlist-header">
                  <Typography variant="h5" sx={{ '&::after': { content: '""', display: 'block', width: 30, height: 2, mt: 0.75, borderRadius: 2, bgcolor: 'primary.main', opacity: 0.55 } }}>Wishlist</Typography>
                </AccordionSummary>
                <AccordionDetails id="wishlist-content">
                  <Stock
                    title="Wishlist"
                    showTitle={false}
                    rows={wishlist}
                    showOwnedColumns={false}
                    createHandler={() => openCreate({ name: '', status: false } as WalletStock)}
                    editHandler={openEdit}
                    deleteHandler={openDelete}
                  />
                </AccordionDetails>
              </Accordion>
            </Stack>
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
