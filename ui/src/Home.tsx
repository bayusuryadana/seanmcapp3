import { useNavigate } from "react-router-dom";
import { Container, Grid, Button } from "@mui/material";

export const Home = () => {

  const navigate = useNavigate()

  return (
    <div id="wrapper">
      <div id="home-container">
        <Container maxWidth="sm">
          <h1>SEANMCAPP</h1>
          <Grid container spacing={3} justifyContent="center">
            <Grid item xs={3}>
              <Button variant="outlined" fullWidth onClick={() => navigate('/wallet')}>Wallet</Button>
            </Grid>
          </Grid>
        </Container>
      </div>
    </div>
  )
}
