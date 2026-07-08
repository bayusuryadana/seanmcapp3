import { createTheme } from '@mui/material/styles';
import { SxProps, Theme } from '@mui/material';

export const drawerWidth: number = 240;

export const defaultTheme = createTheme();

export const modalStyle = {
    position: 'absolute' as const,
    top: '50%',
    left: '50%',
    transform: 'translate(-50%, -50%)',
    width: 400,
    bgcolor: 'background.paper',
    border: '2px solid #000',
    boxShadow: 24,
    p: 4,
  };

// Shared Paper styling used across the dashboards.
export const dashboardPaperStyle: SxProps<Theme> = {
  p: 2,
  display: 'flex',
  flexDirection: 'column',
};

// Shared styling for the main scrollable content area.
export const mainContentBoxStyle: SxProps<Theme> = {
  backgroundColor: (theme) =>
    theme.palette.mode === 'light' ? theme.palette.grey[100] : theme.palette.grey[900],
  flexGrow: 1,
  height: '100vh',
  overflow: 'auto',
};

export const API_URL = import.meta.env.MODE === "development" ? "" : "https://seanmcapp.herokuapp.com"

const stockPoolMoneyEnv = Number(import.meta.env.VITE_STOCK_POOL)
export const STOCK_POOL_MONEY = Number.isFinite(stockPoolMoneyEnv) ? stockPoolMoneyEnv : 0

