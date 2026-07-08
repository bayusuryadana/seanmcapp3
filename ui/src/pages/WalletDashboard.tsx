import { Container, Grid, Paper, Typography, LinearProgress } from "@mui/material"
import { Chart } from "../components/Chart.tsx"
import { Detail } from "../components/Detail.tsx"
import { Title } from "../components/Title.tsx"
import { AppAlert } from "../components/AppAlert.tsx"
import { WalletPlanned, WalletDetail, WalletDashboardData } from "../utils/model.ts"
import { WalletModal } from "../components/Modal.tsx"
import { api } from "../utils/api.ts"
import { useEffect, useState } from "react"
import { dashboardPaperStyle } from "../utils/constant.ts"
import { currentYearMonth } from "../utils/date.ts"
import { useAlert } from "../hooks/useAlert.ts"
import { useModal } from "../hooks/useModal.ts"

export const WalletDashboard = () => {

  const { alert, showError, clearAlert } = useAlert()
  const { modal, openCreate, openEdit, openDelete, close } = useModal<WalletDetail>()
  const [data, setData] = useState<WalletDashboardData|null>(null);
  const [date, setDate] = useState('')

  const onSuccess = () => {
    close()
    if (date !== '') {
      getWalletDashboard(date)
    }
  }

  const getWalletDashboard = (dateParam: string) => {
    api.get('/api/wallet/dashboard', { params: { date: dateParam } })
    .then((response) => {
      clearAlert()
      setData(response.data.data)
      setDate(dateParam)
    })
    .catch(() => showError('Data failed to fetch/parse!'))
  }

  useEffect(() => {
    const dateString = currentYearMonth()
    setDate(dateString)
    getWalletDashboard(dateString)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const totalAllocations = data?.allocations.reduce((acc, item) => {
    acc.expense += item.expense;
    acc.alloc += item.alloc;
    return acc;
  }, { expense: 0, alloc: 0 });

  return (
    <>
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <AppAlert alert={alert} sx={{ mb: 2 }} />
        <Grid container spacing={3}>
          {/* Saving accounts */}
          <Grid item xs={12} md={4}>
            <Paper sx={{ ...dashboardPaperStyle, height: 200, alignItems: 'center' }}>
              <Title>Current Savings</Title>
              <Typography color="text.secondary">
                on DBS account
              </Typography>
              <Typography variant="h6">
                S$ {data ? data.savings.dbs.toLocaleString() : 'Loading...'}
              </Typography>
              <Typography color="text.secondary">
                on BCA account
              </Typography>
              <Typography variant="h6">
                Rp. {data ? data.savings.bca.toLocaleString() : 'Loading...'}
              </Typography>
            </Paper>
          </Grid>
          {/* Balance */}
          <Grid item xs={12} md={8}>
            <Paper sx={{ ...dashboardPaperStyle, height: 200 }}>
                { data?.chart?.balance ? <Chart data={data?.chart.balance} /> : <Typography variant="body2">Loading...</Typography> }
            </Paper>
          </Grid>
          {/* Allocation */}
          <Grid item xs={12} md={4}>
            <Paper sx={{ ...dashboardPaperStyle, alignItems: 'center', px: 3 }}>
              <Title sx={{ mb: 0 }}>Allocations</Title>
              <Typography sx={{ mb: 2 }}>
                ( {totalAllocations?.expense.toLocaleString()} / {totalAllocations?.alloc.toLocaleString()} )
              </Typography>
              <Grid container spacing={2}>
                {data?.allocations.map((item) => {
                  const percent = item.alloc > 0 ? (item.expense / item.alloc) * 100 : 100;
                  return (
                    <Grid item xs={6} md={6} key={item.name}>
                      <Grid container alignItems="baseline" spacing={1} sx={{ mb: 1 }}>
                        <Grid item>
                          <Typography sx={{ fontSize: '0.8rem' }} fontWeight="bold">
                            {item.name}
                          </Typography>
                        </Grid>
                        <Grid item>
                          <Typography sx={{ fontSize: '0.6rem' }} color="text.secondary" ml="3px">
                            ({item.expense.toLocaleString()} / {item.alloc.toLocaleString()})
                          </Typography>
                        </Grid>
                      </Grid>
                      <LinearProgress
                        variant="determinate"
                        value={percent > 100 ? 100 : percent}
                        sx={{
                          height: 10,
                          borderRadius: 5,
                          bgcolor: "grey.300",
                          "& .MuiLinearProgress-bar": {
                            bgcolor: 
                              percent < 60
                              ? "primary.main"
                              : percent < 80
                              ? "#fbc02d"
                              : percent < 100
                              ? "warning.main"
                              : "error.main",
                          },
                        }}
                      />
                    </Grid>
                  );
                })}
              </Grid>
            </Paper>
          </Grid>
          {/* Data */}
          <Grid item xs={12} md={8}>
            <Paper sx={dashboardPaperStyle}>
              <Detail
                date={date}
                rows={data?.detail ?? []} 
                planned={data?.planned ?? { sgd: 0, idr: 0} as WalletPlanned}
                updateDashboard={getWalletDashboard}
                createHandler={() => openCreate()}
                editHandler={openEdit}
                deleteHandler={openDelete}
              />
            </Paper>
          </Grid>
        </Grid>
      </Container>
      
      <WalletModal 
        mode={modal?.mode ?? null}
        detail={modal?.item ?? null}
        date={date}
        onClose={close}
        onSuccess={onSuccess}
      />
    </>
  )
}
