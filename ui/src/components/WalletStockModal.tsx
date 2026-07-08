import { Alert, TextField, Grid, InputLabel, FormControlLabel, Switch, Typography } from "@mui/material";
import { useState, FormEvent, useEffect } from "react";
import { WalletStock } from "../utils/model.ts";
import { api } from "../utils/api.ts";
import { ModalMode, modalTitle } from "../utils/modal.ts";
import { AppAlert } from "./AppAlert.tsx";
import { FormModal } from "./FormModal.tsx";
import { useAlert } from "../hooks/useAlert.ts";

// Stock name must be exactly 4 capital letters (e.g. BBCA)
const STOCK_NAME_REGEX = /^[A-Z]{4}$/

interface Props {
  mode: ModalMode | null
  stock: WalletStock | null
  onClose: () => void
  onSuccess: () => void
}

const parseOptionalNumber = (value: string) => {
  if (value === '') {
    return undefined
  }
  const parsed = parseInt(value)
  return Number.isNaN(parsed) ? undefined : parsed
}

export const WalletStockModal = (props: Props) => {
  const { alert, showError, clearAlert } = useAlert()
  const [data, setData] = useState<WalletStock | null>(null)

  useEffect(() => {
    setData(props.stock)
  }, [props.stock])

  const isOwned = data?.status ?? false
  const totalBought = isOwned && data?.buy_price !== undefined && data?.lot !== undefined
    ? data.buy_price * data.lot * 100
    : undefined

  const validateOwnedFields = () => {
    if (isOwned && ((data?.buy_price ?? 0) <= 0 || (data?.lot ?? 0) <= 0)) {
      showError('Buy Price and Lot are required when Owned is true!')
      return false
    }
    return true
  }

  const validateRequiredPriceFields = () => {
    if ((data?.best_price ?? 0) <= 0 || (data?.fair_price ?? 0) <= 0) {
      showError('Best Price and Fair Price are required and must be > 0!')
      return false
    }
    return true
  }

  const buildPayload = (name: string) => ({
    name,
    best_price: data?.best_price,
    fair_price: data?.fair_price,
    status: isOwned,
    buy_price: isOwned ? (data?.buy_price ?? null) : null,
    lot: isOwned ? (data?.lot ?? null) : null,
  })

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (props.mode === 'delete') {
      submitDelete()
    } else {
      submitUpsert()
    }
  }

  const submitUpsert = () => {
    const isEdit = props.mode === 'edit'
    const name = isEdit ? (props.stock?.name ?? "") : (data?.name ?? "")
    if (!isEdit && !STOCK_NAME_REGEX.test(name)) {
      showError('Name must be exactly 4 capital letters!')
      return
    }
    if (!validateOwnedFields() || !validateRequiredPriceFields()) {
      return
    }
    const url = isEdit ? '/api/stock/update' : '/api/stock/create'

    api.post(url, buildPayload(name))
      .then(() => {
        clearAlert()
        props.onSuccess()
      })
      .catch(() => showError(isEdit ? 'Failed to update!' : 'Failed to create!'))
  }

  const submitDelete = () => {
    const name = props.stock?.name ?? ''
    api.delete('/api/stock/delete/' + name)
      .then((response) => {
        if (response.data.data === name) {
          props.onSuccess()
        } else {
          showError('something is wrong with the API')
        }
      })
      .catch(() => showError('Failed to delete!'))
  }

  const renderForm = () => {
    return (
      <>
        <AppAlert alert={alert} sx={{ mb: 1 }} />
        <Grid container spacing={1}>
          <Grid item xs={12}>
            <InputLabel>Name</InputLabel>
            <TextField
              required fullWidth name="name" type="text" variant="standard"
              value={data?.name ?? ''}
              disabled={props.mode === 'edit'}
              helperText="4 capital letters (e.g. BBCA)"
              inputProps={{ pattern: '[A-Z]{4}', maxLength: 4, style: { textTransform: 'uppercase' } }}
              onChange={(event) => setData({...data, name: event.target.value.toUpperCase()} as WalletStock)}
            />
          </Grid>
          <Grid item xs={6}>
            <InputLabel>Best Price</InputLabel>
            <TextField
              required fullWidth name="best_price" type="number" variant="standard"
              value={data?.best_price ?? ''}
              onChange={(event) => setData({...data, best_price: parseOptionalNumber(event.target.value)} as WalletStock)}
            />
          </Grid>
          <Grid item xs={6}>
            <InputLabel>Fair Price</InputLabel>
            <TextField
              required fullWidth name="fair_price" type="number" variant="standard"
              value={data?.fair_price ?? ''}
              onChange={(event) => setData({...data, fair_price: parseOptionalNumber(event.target.value)} as WalletStock)}
            />
          </Grid>
          <Grid item xs={12}>
            <FormControlLabel
              control={
                <Switch
                  color="secondary"
                  name="status"
                  checked={data?.status ?? false}
                  onChange={(event) => setData({
                    ...data,
                    status: event.target.checked,
                    buy_price: event.target.checked ? data?.buy_price : undefined,
                    lot: event.target.checked ? data?.lot : undefined,
                  } as WalletStock)}
                />
              }
              label="Owned?"
            />
          </Grid>
          {isOwned && (
            <>
              <Grid item xs={6}>
                <InputLabel>Buy Price</InputLabel>
                <TextField
                  required
                  fullWidth
                  name="buy_price"
                  type="number"
                  variant="standard"
                  value={data?.buy_price ?? ''}
                  onChange={(event) => setData({...data, buy_price: parseOptionalNumber(event.target.value)} as WalletStock)}
                  helperText="Required when Owned is true"
                />
              </Grid>
              <Grid item xs={6}>
                <InputLabel>Lot</InputLabel>
                <TextField
                  required
                  fullWidth
                  name="lot"
                  type="number"
                  variant="standard"
                  value={data?.lot ?? ''}
                  onChange={(event) => setData({...data, lot: parseOptionalNumber(event.target.value)} as WalletStock)}
                  helperText="1 lot = 100 shares"
                />
              </Grid>
              <Grid item xs={12}>
                <Typography variant="caption" color="text.secondary">
                  Total bought = Buy Price x Lot x 100 {totalBought !== undefined ? `= ${totalBought.toLocaleString()}` : ''}
                </Typography>
              </Grid>
            </>
          )}
        </Grid>
      </>
    )
  }

  const isDelete = props.mode === 'delete'

  return (
    <FormModal
      open={props.mode !== null}
      title={`${props.mode ? modalTitle[props.mode] : ''} Stock`}
      submitLabel={isDelete ? 'Delete' : 'Submit'}
      onClose={props.onClose}
      onSubmit={handleSubmit}
    >
      {isDelete ? (
        <>
          <AppAlert alert={alert} sx={{ mb: 1 }} />
          <Alert severity="warning">Are you sure you want to delete <b>{props.stock?.name}</b>?</Alert>
        </>
      ) : renderForm()}
    </FormModal>
  );
}
