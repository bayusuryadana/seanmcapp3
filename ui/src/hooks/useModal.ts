import { useState } from "react"
import { ModalMode } from "../utils/modal"

export type ModalState<T> = {
  mode: ModalMode
  item: T | null
}

// Manages create/edit/delete dialog state for a given item type.
export const useModal = <T,>() => {
  const [modal, setModal] = useState<ModalState<T> | null>(null)

  return {
    modal,
    openCreate: (item: T | null = null) => setModal({ mode: "create", item }),
    openEdit: (item: T) => setModal({ mode: "edit", item }),
    openDelete: (item: T) => setModal({ mode: "delete", item }),
    close: () => setModal(null),
  }
}

