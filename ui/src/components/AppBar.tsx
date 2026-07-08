import MuiAppBar from '@mui/material/AppBar';
import { styled } from '@mui/material/styles';
import { Toolbar, IconButton, Typography, Button, Box } from '@mui/material';
import LogoutIcon from '@mui/icons-material/Logout';
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
    <AppBar position="absolute">
      <Toolbar sx={{ pr: '24px', }}>
        <Typography component="h1" variant="h6" color="inherit" noWrap sx={{ mr: 3 }}>
          Seanmcwallet
        </Typography>
        <Box sx={{ flexGrow: 1, display: 'flex', gap: 1 }}>
          {navItems.map((item) => (
            <Button
              key={item.path}
              color="inherit"
              onClick={() => navigate(item.path)}
              sx={{
                fontWeight: location.pathname === item.path ? 'bold' : 'normal',
                borderBottom: location.pathname === item.path ? '2px solid' : '2px solid transparent',
                borderRadius: 0,
              }}
            >
              {item.label}
            </Button>
          ))}
        </Box>
        <IconButton color="inherit" onClick={props.logoutHandler}>
          <LogoutIcon />
        </IconButton>
      </Toolbar>
    </AppBar>
  )
}