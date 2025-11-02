import { Container, Alert, Grid, Paper, Typography, LinearProgress } from "@mui/material"
import Chart from "../components/Chart.tsx"
import { Detail } from "../components/Detail.tsx"
import { Title } from "../components/Title.tsx"
import { WalletPlanned, WalletDetail, WalletDashboardData, WalletAlert } from "../utils/model.ts"
import { WalletModal } from "../components/Modal.tsx"
import axios from "axios"
import { useContext, useEffect, useState } from "react"
import { UserContext, UserContextType } from "../UserContext.tsx"
import { API_URL } from "../utils/constant.ts"

export const WalletDashboard = () => {

  const { userContext, saveToken } = useContext(UserContext) as UserContextType
  const [alert, setAlert] = useState<WalletAlert>({display: 'none', text: ''})
  const [data, setData] = useState<WalletDashboardData|null>(null);
  const [walletDetail, setWalletDetail] = useState<WalletDetail|null>(null)
  const [date, setDate] = useState('')

  const onSuccess = (row: WalletDetail, actionText: String|undefined) => {
    setWalletDetail(null)
    if (data !== null) {
      if (actionText === 'Create') {
        const updatedDetail = {...data, detail: [...data.detail, row]}
        setData(updatedDetail)
      } else if (actionText === 'Edit') {
        const index = data?.detail.findIndex((d) => d.id === row.id) ?? -1
        if (index && index > -1 && data) {
          const updatedDetail = {...data, detail: [...data.detail.filter((_, i) => i !== index), row]}
          setData(updatedDetail)
        }
      } else if (actionText === 'Delete') {
        const index = data?.detail.findIndex((d) => d.id === row.id) ?? -1
        if (index && index > -1 && data) {
          setData({...data, detail: data.detail.filter((_, i) => i !== index)})
        }
      }
    }
  }

  const getWalletDashboard = (dateParam: string) => {
    axios.get(API_URL + '/api/wallet/dashboard', {
      headers: {
        Authorization: 'Bearer ' + (userContext ?? "")
      },
      params: {
        date: dateParam
      },
    })
    .then((response) => {
      setAlert({display: 'none', text: ''})
      setData(response.data.data)
      setDate(dateParam)
    })
    .catch((error) => {
      console.log(error)
      if (axios.isAxiosError(error) && error.response?.status == 401) {
        saveToken(null)
      } else {
        setAlert({display: 'true', text: 'Data failed to fetch/parse!'})
      }
    })
  }

  useEffect(() => {
    const newDate = new Date()
    const dateString = newDate.getFullYear().toString() + ('0' + (newDate.getMonth() + 1).toString()).slice(-2)
    setDate(dateString)
    getWalletDashboard(dateString)
  }, [])

  const totalAllocations = data?.allocations.reduce((acc, item) => {
    acc.expense += item.expense;
    acc.alloc += item.alloc;
    return acc;
  }, { expense: 0, alloc: 0 });

  return (
    <>
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Alert id="invalid-data-alert" severity="error" sx={{ mb: 2, display: alert.display}}>{alert.text}</Alert>
        <Grid container spacing={3}>
          {/* Saving accounts */}
          <Grid item xs={12} md={4}>
            <Paper sx={{p: 2, display: 'flex', flexDirection: 'column', height: 200, alignItems: 'center', }}>
              <Title>Current Savings</Title>
              <Typography color="text.secondary">
                on DBS account
              </Typography>
              <Typography variant="h6">
                S$ {data?.savings?.dbs ? data?.savings.dbs.toLocaleString() : 'Loading...'}
              </Typography>
              <Typography color="text.secondary">
                on BCA account
              </Typography>
              <Typography variant="h6">
                Rp. {data?.savings?.bca ? data?.savings.bca.toLocaleString() : 'Loading...'}
              </Typography>
            </Paper>
          </Grid>
          {/* Balance */}
          <Grid item xs={12} md={8}>
            <Paper sx={{p: 2, display: 'flex', flexDirection: 'column', height: 200, }}>
                { data?.chart?.balance ? <Chart data={data?.chart.balance} /> : <Typography variant="body2">Loading...</Typography> }
            </Paper>
          </Grid>
          {/* Allocation */}
          <Grid item xs={12} md={4}>
            <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column', alignItems: 'center', px: 3 }}>
              <Title sx={{ mb: 0 }}>Allocations</Title>
              <Typography sx={{ mb: 2 }}>
                ( {totalAllocations?.expense.toLocaleString()} / {totalAllocations?.alloc.toLocaleString()} )
              </Typography>
              <Grid container spacing={2}>
                {data?.allocations.map((item) => {
                  const percent = item.alloc > 0 ? (item.expense / item.alloc) * 100 : 100;
                  console.log(item.name + ": " + percent)
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
            <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
              <Detail 
                date={date}
                rows={data?.detail ?? []} 
                planned={data?.planned ?? { sgd: 0, idr: 0} as WalletPlanned}
                updateDashboard={getWalletDashboard}
                createHandler={() => {setWalletDetail({ id: -1 } as WalletDetail)}}
                editHandler={(walletDetail: WalletDetail) => {setWalletDetail(walletDetail)}} 
                deleteHandler={(id: Number) => {setWalletDetail({ id: id } as WalletDetail)}} 
              />
            </Paper>
          </Grid>
        </Grid>
      </Container>
      
      <WalletModal 
          onClose={() => setWalletDetail(null)}
          date={date}
          onSuccess={onSuccess}
          walletDetail={walletDetail}
          />
    </>
  )
}