import { Title } from './Title';
import { WalletDetail, WalletPlanned } from '../utils/model';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import { Grid, IconButton, TableRow, TableHead, TableCell, TableBody, Table, Button, Popover, Box, TextField, Typography, TableContainer } from '@mui/material';
import ArrowLeftIcon from '@mui/icons-material/ArrowLeft';
import ArrowRightIcon from '@mui/icons-material/ArrowRight';
import { FormEvent, Fragment, useState } from 'react';
import { CellTypography } from './CellTypography';
import { AppAlert } from './AppAlert';
import { useAlert } from '../hooks/useAlert';
import { isValidYearMonth, shiftYearMonth, yearMonthTitle } from '../utils/date';

interface DetailProps {
  date: string
  rows: WalletDetail[]
  planned: WalletPlanned
  editHandler: (row: WalletDetail) => void
  deleteHandler: (row: WalletDetail) => void
  createHandler: () => void
  updateDashboard: (date: string) => void
}

export const Detail = (props: DetailProps) => {

  const [anchorEl, setAnchorEl] = useState<HTMLButtonElement | null>(null);
  const { alert, showError, clearAlert } = useAlert()

  const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const open = Boolean(anchorEl);
  const id = open ? 'simple-popover' : undefined;

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const input = data.get('dateInput')?.toString() ?? ""
    if (!isValidYearMonth(input)) {
      showError('Salah format goblok!')
    } else {
      props.updateDashboard(input)
      clearAlert()
      handleClose()
    }
  }

  return (
    <Fragment>
      <Grid container justifyContent={'space-between'}>
        <Grid item>
          <Title>Detail</Title>
        </Grid>
        <Grid item>
          <IconButton color='primary' size='medium' sx={{display: 'inline'}} onClick={() => props.updateDashboard(shiftYearMonth(props.date, -1))}>
            <ArrowLeftIcon />
          </IconButton>
          <Button aria-describedby={id} variant="contained" onClick={handleClick}>
            {yearMonthTitle(props.date)}
          </Button>
          <Popover
            id={id}
            open={open}
            anchorEl={anchorEl}
            onClose={handleClose}
            anchorOrigin={{
              vertical: 'bottom',
              horizontal: 'left',
            }}
          >
            <Box component="form" onSubmit={handleSubmit} noValidate sx={{ mt: 1, p: 2 }}>
              <AppAlert alert={alert} />
              <TextField margin="normal" required fullWidth name="dateInput" label="Which month you want?" type="number" id="dateInput"/>
              <Button type="submit" fullWidth variant="contained" sx={{ mt: 3, mb: 2 }}>
                GO!
              </Button>
            </Box> 
          </Popover>
          <IconButton color='primary' size='medium' sx={{display: 'inline'}} onClick={() => props.updateDashboard(shiftYearMonth(props.date, 1))}>
            <ArrowRightIcon />
          </IconButton>
        </Grid>
        <Grid item>
          <IconButton color='primary' size='small' onClick={props.createHandler}>
            <AddIcon />
          </IconButton>
        </Grid>
      </Grid>
      <TableContainer sx={{ overflowX: 'auto' }}>
        <Table size="small" sx={{ '& td, & th': { px: 0.5, py: 0.5, fontSize: '0.75rem' } }}>
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>Category</TableCell>
              <TableCell>Currency</TableCell>
              <TableCell align="right">Amount</TableCell>
              <TableCell></TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {props.rows.map((row) => (
              <TableRow key={row.id}>
                <TableCell><CellTypography done={row.done}>{row.name}</CellTypography></TableCell>
                <TableCell><CellTypography done={row.done}>{row.category}</CellTypography></TableCell>
                <TableCell><CellTypography done={row.done}>{row.currency}</CellTypography></TableCell>
                <TableCell align="right"><CellTypography done={row.done}>{row.amount.toLocaleString()}</CellTypography></TableCell>
                <TableCell sx={{ whiteSpace: "nowrap" }}>
                  <IconButton size="small" sx={{ p: 0.25 }} aria-label="edit" color="primary" onClick={()=>props.editHandler(row)}>
                    <EditIcon fontSize="small" />
                  </IconButton>
                  <IconButton size="small" sx={{ p: 0.25 }} aria-label="delete" color="secondary" onClick={()=>props.deleteHandler(row)}>
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      <Typography sx={{ p: 2, fontSize: { xs: "0.75rem", sm: "0.875rem" }, }}>
        Cash balance end of month | SGD: <b>S$ {props.planned.sgd.toLocaleString()}</b> | IDR: <b>Rp. {props.planned.idr.toLocaleString()}</b>
      </Typography>
    </Fragment>
  );
}
