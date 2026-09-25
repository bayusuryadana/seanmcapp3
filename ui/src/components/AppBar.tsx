import MuiAppBar from '@mui/material/AppBar';
import { styled } from '@mui/material/styles';
import { Toolbar, IconButton, Typography, Button, Box, BottomNavigation, BottomNavigationAction } from '@mui/material';
import LogoutIcon from '@mui/icons-material/Logout';
import DashboardIcon from '@mui/icons-material/Dashboard';
import ShowChartIcon from '@mui/icons-material/ShowChart';
import { useNavigate, useLocation } from 'react-router-dom';

export const AppBar = styled(MuiAppBar)(({ theme }) => ({
  zIndex: theme.zIndex.drawer + 1,
}));

interface WalletAppBarProps {
  logoutHandler: () => void
}

export const WalletAppBar = (props: WalletAppBarProps) => {
  const navigate = useNavigate()
  const location = useLocation()

  const navItems = [
    { label: 'Dashboard', path: '/wallet' },
    { label: 'Stock', path: '/wallet/stock' },
  ]

  return (
    <>
    <AppBar position="fixed" elevation={1}>
      <Toolbar sx={{ pr: { xs: 1, sm: 3 }, pl: { xs: 2, sm: 3 }, pt: 'env(safe-area-inset-top)' }}>
        <Typography component="h1" variant="h6" color="inherit" noWrap sx={{ mr: { xs: 1, sm: 3 }, fontSize: { xs: '1rem', sm: '1.25rem' } }}>
          Seanmcwallet
        </Typography>
        <Box sx={{ flexGrow: 1, display: 'flex', gap: 1 }}>
          {navItems.map((item) => (
            <Button
              key={item.path}
              color="inherit"
              onClick={() => navigate(item.path)}
              sx={{
                display: { xs: 'none', sm: 'inline-flex' },
                fontWeight: location.pathname === item.path ? 'bold' : 'normal',
                borderBottom: location.pathname === item.path ? '2px solid' : '2px solid transparent',
                borderRadius: 0,
              }}
            >
              {item.label}
            </Button>
          ))}
        </Box>
        <IconButton color="inherit" aria-label="Log out" onClick={props.logoutHandler}>
          <LogoutIcon />
        </IconButton>
      </Toolbar>
    </AppBar>
    <BottomNavigation
      value={location.pathname}
      showLabels
      sx={{
        display: { xs: 'flex', sm: 'none' },
        position: 'fixed', bottom: 0, left: 0, right: 0, zIndex: (theme) => theme.zIndex.appBar,
        height: 'calc(56px + env(safe-area-inset-bottom))', pb: 'env(safe-area-inset-bottom)',
        borderTop: 1, borderColor: 'divider',
      }}
    >
      <BottomNavigationAction aria-label="Dashboard tab" label="Dashboard" value="/wallet" icon={<DashboardIcon />} onClick={() => navigate('/wallet')} />
      <BottomNavigationAction aria-label="Stock tab" label="Stock" value="/wallet/stock" icon={<ShowChartIcon />} onClick={() => navigate('/wallet/stock')} />
    </BottomNavigation>
    </>
  )
}
