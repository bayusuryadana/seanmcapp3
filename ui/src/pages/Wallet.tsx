import { ThemeProvider } from '@mui/material/styles';
import { defaultTheme, mainContentBoxStyle } from '../utils/constant';
import { Navigate, Outlet } from 'react-router-dom';
import { useUser } from '../hooks/useUser';
import { Box, CssBaseline, Toolbar } from '@mui/material';
import { WalletAppBar } from '../components/AppBar';

export const Wallet = () => {
    const { userContext, saveToken } = useUser();
    const logoutHandler = () => saveToken(null)

    if (userContext != null) {
      return (
        <ThemeProvider theme={defaultTheme}>
          <Box sx={{ display: 'flex' }}>
            <CssBaseline />
            <WalletAppBar logoutHandler={logoutHandler} />
            <Box component="main" sx={mainContentBoxStyle}>
              <Toolbar sx={{ minHeight: { xs: 'calc(56px + env(safe-area-inset-top)) !important', sm: '64px !important' } }} />
              <Outlet />
            </Box>
          </Box>
        </ThemeProvider>
      );
    } else {
      return <Navigate to="/wallet/login" />
    }
}
