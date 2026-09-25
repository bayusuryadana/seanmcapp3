import { defaultTheme, mainContentBoxStyle } from "../utils/constant.ts"
import CssBaseline from '@mui/material/CssBaseline';
import { AppBar } from '../components/AppBar.tsx';
import { AppAlert } from '../components/AppAlert.tsx';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import { FormEvent } from 'react';
import { Navigate } from "react-router-dom";
import { api } from "../utils/api.ts";
import axios from "axios";
import { Paper, Avatar, Button, ThemeProvider, Box, Toolbar, Typography, TextField } from "@mui/material";
import { useUser } from "../hooks/useUser.ts";
import { useAlert } from "../hooks/useAlert.ts";

export const WalletLogin = () => {
  const { userContext, saveToken } = useUser();
  const { alert, showError, clearAlert } = useAlert()

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const inputPassword = data.get('password')?.toString() ?? ""

    api.post('/api/wallet/login', { password: inputPassword })
    .then((response) => {
      clearAlert()
      saveToken(response.data)
    })
    .catch((error) => {
      const status = axios.isAxiosError(error) ? error.response?.status : undefined
      if (status === 401 || status === 403) {
        showError('Salah password goblok!')
      } else {
        showError('Gatau nih gabisanya kenapa tot!')
      }
    });
  };

  if (userContext == null) {
    return (
      <ThemeProvider theme={defaultTheme}>
        <Box sx={{ display: 'flex' }}>
          <CssBaseline />
          <AppBar position="fixed">
            <Toolbar sx={{ pr: { xs: 2, sm: 3 }, pt: 'env(safe-area-inset-top)' }}>
              <Typography component="h1" variant="h6" color="inherit" noWrap sx={{ flexGrow: 1 }}>
                Seanmcwallet
              </Typography>
            </Toolbar>
          </AppBar>

          <Box component="main" sx={{
              ...mainContentBoxStyle,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              minHeight: { xs: 'calc(100dvh - 56px - env(safe-area-inset-top))', sm: 'calc(100dvh - 64px)' },
              marginTop: { xs: 'calc(56px + env(safe-area-inset-top))', sm: 8 },
              padding: { xs: 2, sm: 4 },
              pt: { xs: 8, sm: 12 },
              justifyContent: 'flex-start',
            }}>
              <Paper sx={{ p: { xs: 2.5, sm: 3 }, width: '100%', maxWidth: 400, display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                <Avatar sx={{ m: 1, bgcolor: 'secondary.main' }}>
                  <LockOutlinedIcon />
                </Avatar>
                <Typography component="h1" variant="h5">
                  Sign in
                </Typography>
                <Box component="form" onSubmit={handleSubmit} noValidate sx={{ mt: 1 }}>
                  <AppAlert alert={alert} />
                  <TextField margin="normal" required fullWidth name="password" label="Password" type="password" id="password" autoComplete="current-password" />
                  <Button type="submit" fullWidth variant="contained" sx={{ mt: 3, mb: 2 }}>
                      Sign In
                  </Button>
                </Box> 
            </Paper>
          </Box>
        </Box>
      </ThemeProvider>
    );
  } else {
    return <Navigate to="/wallet" />
  }
}
