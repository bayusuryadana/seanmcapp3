import { TextField, MenuItem, Select, Grid, InputLabel, FormControlLabel, Checkbox } from "@mui/material";
import { useState, FormEvent, useEffect } from "react";
import { WalletDetail } from "../utils/model.ts";
import { api } from "../utils/api.ts";
import { ModalMode, modalTitle } from "../utils/modal.ts";
import { AppAlert } from "./AppAlert.tsx";
import { FormModal } from "./FormModal.tsx";
import { useAlert } from "../hooks/useAlert.ts";

interface WalletModalProps {
  mode: ModalMode | null
  detail: WalletDetail | null
  date: string
  onClose: () => void
  onSuccess: () => void
}

const CATEGORIES = [
  'Bonus', 'Daily', 'Fashion', 'Funding', 'IT Stuff', 'Misc', 'ROI',
  'Rent', 'Salary', 'Temp', 'Transfer', 'Travel', 'Wellness', 'Zakat',
]

const getAccount = (currency: string): string => {
  if (currency === 'SGD') return 'DBS'
  if (currency === 'IDR') return 'BCA'
  return ''
}

export const WalletModal = (props: WalletModalProps) => {
  const { alert, showError, clearAlert } = useAlert()
  const [data, setData] = useState<WalletDetail | null>(null)

  useEffect(() => {
    setData(props.mode === 'create' ? null : props.detail)
  }, [props.mode, props.detail])

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
    const currency = data?.currency ?? ""
    const payload = {
      ...(isEdit ? { id: props.detail?.id } : {}),
      date: parseInt(props.date),
      name: data?.name ?? "",
      amount: data?.amount ?? 0,
      category: data?.category ?? "",
      currency,
      account: getAccount(currency),
      done: data?.done ?? false,
    }
    const url = isEdit ? '/api/wallet/update' : '/api/wallet/create'

    api.post(url, payload)
      .then(() => {
        clearAlert()
        props.onSuccess()
      })
      .catch(() => showError('Gagal tot!'))
  }

  const submitDelete = () => {
    const id = props.detail?.id ?? -1
    api.delete('/api/wallet/delete/' + id)
      .then((response) => {
        if (response.data.data === id) {
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
            <TextField required fullWidth name="name" type="text" value={data?.name ?? ''} variant="standard" onChange={(event) => setData({...data, name: event.target.value} as WalletDetail)} />
          </Grid>
          <Grid item xs={12}>
            <InputLabel>Amount</InputLabel>
            <TextField required fullWidth name="amount" type="number" value={data?.amount ?? ''} variant="standard" onChange={(event) => setData({...data, amount: parseInt(event.target.value)} as WalletDetail)} />
          </Grid>
          <Grid item xs={12}>
            <InputLabel>Category</InputLabel>
            <Select
              required
              fullWidth
              value={data?.category ?? ""}
              label="Category"
              name="category"
              variant="standard"
              onChange={(event) => setData({...data, category: event.target.value} as WalletDetail)}
            >
              {CATEGORIES.map((category) => (
                <MenuItem key={category} value={category}>{category}</MenuItem>
              ))}
            </Select>
          </Grid>
          <Grid item xs={6}>
            <InputLabel>Currency</InputLabel>
            <Select
              required
              fullWidth
              value={data?.currency ?? ""}
              label="currency"
              name="currency"
              variant="standard"
              onChange={(event) => {
                const currency = event.target.value
                setData({...data, currency, account: getAccount(currency)} as WalletDetail)
              }}
            >
              <MenuItem value={'SGD'}>SGD</MenuItem>
              <MenuItem value={'IDR'}>IDR</MenuItem>
            </Select>
          </Grid>
          <Grid item xs={6}>
            <InputLabel>Account</InputLabel>
            <TextField required disabled fullWidth value={data?.account ?? ''} name="account" type="text" variant="standard"/>
          </Grid>
          <Grid item xs={12}>
            <FormControlLabel
              control={
                <Checkbox
                  color="secondary"
                  name="done"
                  checked={data?.done ?? false}
                  onChange={(event) => setData({...data, done: event.target.checked} as WalletDetail)} />
              }
              label="Is it done?"
            />
          </Grid>
        </Grid>
      </>
    )
  }

  const isDelete = props.mode === 'delete'

  return (
    <FormModal
      open={props.mode !== null}
      title={props.mode ? modalTitle[props.mode] : ''}
      submitLabel={isDelete ? 'Delete' : 'Submit'}
      onClose={props.onClose}
      onSubmit={handleSubmit}
    >
      {isDelete || renderForm()}
    </FormModal>
  );
}
