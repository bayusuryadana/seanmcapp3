import {WalletStock} from "../utils/model.ts";

interface Props {
    onClose: () => void
    onSuccess: (row: WalletStock, actionText: string|undefined) => void
    walletStock: WalletStock|null
}

export const WalletStockModal = (props: Props) => {
    return (<>{props}</>);
}