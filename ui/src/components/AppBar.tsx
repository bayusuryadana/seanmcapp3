import MuiAppBar, { AppBarProps as MuiAppBarProps } from '@mui/material/AppBar';
import { styled } from '@mui/material/styles';
import { drawerWidth } from '../utils/constant.ts';
import { Toolbar, IconButton, Typography, Button, Box } from '@mui/material';
import LogoutIcon from '@mui/icons-material/Logout';
import { useNavigate, useLocation } from 'react-router-dom';

interface AppBarProps extends MuiAppBarProps {
  open?: boolean;
}
  
export const AppBar = styled(MuiAppBar, {
  shouldForwardProp: (prop) => prop !== 'open',
})<AppBarProps>(({ theme, open }) => ({
  zIndex: theme.zIndex.drawer + 1,
  transition: theme.transitions.create(['width', 'margin'], {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
  ...(open && {
  marginLeft: drawerWidth,
  width: `calc(100% - ${drawerWidth}px)`,
  transition: theme.transitions.create(['width', 'margin'], {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.enteringScreen,
  }),
  }),
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
    <AppBar position="absolute" open={false}>
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