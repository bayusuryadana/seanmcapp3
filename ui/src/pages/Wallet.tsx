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
              <Toolbar />{/* this Toolbar is supposed for bufffer below the real AppBar */}
              <Outlet />
            </Box>
          </Box>
        </ThemeProvider>
      );
    } else {
      return <Navigate to="/wallet/login" />
    }
}
