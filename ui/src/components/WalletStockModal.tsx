import { Button, Modal, Box, Typography, Alert, TextField, Grid, InputLabel, FormControlLabel, Switch } from "@mui/material";
import axios from "axios";
import { useContext, useState, FormEvent, useEffect } from "react";
import { UserContext, UserContextType } from "../UserContext.tsx";
import { WalletStock } from "../utils/model.ts";
import { API_URL, modalStyle } from "../utils/constant.ts";

// Stock name must be exactly 4 capital letters (e.g. BBCA)
const STOCK_NAME_REGEX = /^[A-Z]{4}$/

interface Props {
    onClose: () => void
    onSuccess: () => void
    walletStock: WalletStock|null
}

export const WalletStockModal = (props: Props) => {
    const { userContext } = useContext(UserContext) as UserContextType;
    const [alert, setAlert] = useState({display: 'none', text: ''})
    const [data, setData] = useState<WalletStock|null>(null)

    useEffect(() => {
        setData(props.walletStock)
    }, [props.walletStock])

    const getActionText = () => {
        const stock = props.walletStock ?? {} as WalletStock
        if (stock.name === '') {
            return 'Create'
        } else if (stock.name &&
            stock.best_price !== undefined && stock.fair_price !== undefined && stock.status !== undefined) {
            return 'Edit'
        } else if (stock.name) {
            return 'Delete'
        }
    }
    const actionText = getActionText()

    const parseOptionalNumber = (value: string) => {
        if (value === '') {
            return undefined
        }
        const parsed = parseInt(value)
        return Number.isNaN(parsed) ? undefined : parsed
    }

    const isOwned = data?.status ?? false
    const totalBought = isOwned && data?.buy_price !== undefined && data?.lot !== undefined
        ? data.buy_price * data.lot * 100
        : undefined

    const validateOwnedFields = () => {
        if (isOwned && ((data?.buy_price ?? 0) <= 0 || (data?.lot ?? 0) <= 0)) {
            setAlert({display: 'true', text: 'Buy Price and Lot are required when Owned is true!'})
            return false
        }
        return true
    }

    const validateRequiredPriceFields = () => {
        if ((data?.best_price ?? 0) <= 0 || (data?.fair_price ?? 0) <= 0) {
            setAlert({display: 'true', text: 'Best Price and Fair Price are required and must be > 0!'})
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
        if (actionText === 'Create') {
            submitCreate()
        } else if (actionText === 'Edit') {
            submitEdit()
        } else if (actionText === 'Delete') {
            submitDelete()
        }
    }

    const submitCreate = () => {
        const name = data?.name ?? ""
        if (!STOCK_NAME_REGEX.test(name)) {
            setAlert({display: 'true', text: 'Name must be exactly 4 capital letters!'})
            return
        }
        if (!validateOwnedFields()) {
            return
        }
        if (!validateRequiredPriceFields()) {
            return
        }
        const payload = buildPayload(name)

        axios.post(API_URL + '/api/stock/create', payload, {
            headers: { Authorization: 'Bearer ' + (userContext ?? "") }
        }).then(() => {
            setAlert({display: 'none', text: ''})
            props.onSuccess()
        }).catch((error) => {
            console.log(error)
            setAlert({display: 'true', text: 'Failed to create!'})
        })
    }

    const submitEdit = () => {
        if (!validateOwnedFields()) {
            return
        }
        if (!validateRequiredPriceFields()) {
            return
        }
        const payload = buildPayload(props.walletStock?.name ?? "")

        axios.post(API_URL + '/api/stock/update', payload, {
            headers: { Authorization: 'Bearer ' + (userContext ?? "") }
        }).then(() => {
            setAlert({display: 'none', text: ''})
            props.onSuccess()
        }).catch((error) => {
            console.log(error)
            setAlert({display: 'true', text: 'Failed to update!'})
        })
    }

    const submitDelete = () => {
        const name = props.walletStock?.name ?? ''
        axios.delete(API_URL + '/api/stock/delete/' + name, {
            headers: { Authorization: 'Bearer ' + (userContext ?? "") }
        }).then((response) => {
            if (response.data.data === name) {
                props.onSuccess()
            } else {
                const errorMessage = 'something is wrong with the API'
                console.log(errorMessage)
                setAlert({display: 'true', text: errorMessage})
            }
        }).catch((error) => {
            console.log(error)
            setAlert({display: 'true', text: 'Failed to delete!'})
        })
    }

    const renderForm = () => {
        return (
            <>
                <Alert severity="error" sx={{display: alert.display, mb: 1}}>{alert.text}</Alert>
                <Grid container spacing={1}>
                    <Grid item xs={12}>
                        <InputLabel>Name</InputLabel>
                        <TextField
                            required fullWidth name="name" type="text" variant="standard"
                            value={data?.name ?? ''}
                            disabled={actionText === 'Edit'}
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
                                    value={data?.status ? 'yes' : ''}
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
                                    required={isOwned}
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
                                    required={isOwned}
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

    return (
        <Modal
            open={props.walletStock !== null}
            onClose={props.onClose}
            aria-labelledby="stock-modal-title"
            aria-describedby="stock-modal-description"
        >
            <Box sx={modalStyle}>
                <Typography id="stock-modal-title" variant="h6" component="h2">
                    {actionText} Stock
                </Typography>
                <Box component="form" onSubmit={handleSubmit} sx={{ mt: 2 }}>
                    {actionText === 'Delete' ? (
                        <Alert severity="warning">Are you sure you want to delete <b>{props.walletStock?.name}</b>?</Alert>
                    ) : renderForm()}
                    <Button type="submit" fullWidth variant="contained" sx={{ mt: 3, mb: 2 }}>
                        {actionText === 'Delete' ? 'Delete' : 'Submit'}
                    </Button>
                </Box>
            </Box>
        </Modal>
    );
}
