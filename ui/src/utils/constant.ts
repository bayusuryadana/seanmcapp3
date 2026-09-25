import { createTheme } from '@mui/material/styles';
import { SxProps, Theme } from '@mui/material';


export const defaultTheme = createTheme({
  palette: {
    primary: { main: '#1565c0' },
    background: { default: '#f5f7fa' },
  },
  shape: { borderRadius: 12 },
  components: {
    MuiButton: { defaultProps: { disableElevation: true } },
    MuiPaper: { defaultProps: { elevation: 1 } },
  },
});

export const modalStyle = {
    position: 'absolute' as const,
    top: '50%',
    left: '50%',
    transform: 'translate(-50%, -50%)',
    width: { xs: 'calc(100% - 32px)', sm: 400 },
    maxHeight: 'calc(100dvh - 32px)',
    overflowY: 'auto',
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
  minHeight: '100dvh',
  overflow: 'auto',
  pb: { xs: 'calc(56px + env(safe-area-inset-bottom))', sm: 0 },
};

// Shared styling for the compact data tables.
export const tableContainerStyle: SxProps<Theme> = { overflowX: 'auto' };
export const compactTableStyle: SxProps<Theme> = {
  '& td, & th': { px: 0.5, py: 0.5, fontSize: '0.75rem' },
};

export const API_URL = import.meta.env.MODE === "development" ? "" : "https://seanmcapp.herokuapp.com"

const stockPoolMoneyEnv = Number(import.meta.env.VITE_STOCK_POOL)
export const STOCK_POOL_MONEY = Number.isFinite(stockPoolMoneyEnv) ? stockPoolMoneyEnv : 0
