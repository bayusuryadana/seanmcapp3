import { createTheme } from '@mui/material/styles';

export const drawerWidth: number = 240;

export const defaultTheme = createTheme();

export const modalStyle = {
    position: 'absolute' as 'absolute',
    top: '50%',
    left: '50%',
    transform: 'translate(-50%, -50%)',
    width: 400,
    bgcolor: 'background.paper',
    border: '2px solid #000',
    boxShadow: 24,
    p: 4,
  };

export const API_URL = import.meta.env.MODE === "development" ? "http://localhost:8080" : "https://seanmcapp.herokuapp.com"
