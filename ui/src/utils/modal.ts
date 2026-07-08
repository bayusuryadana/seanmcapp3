// Shared modal mode used by the create/edit/delete dialogs.
export type ModalMode = "create" | "edit" | "delete"

export const modalTitle: Record<ModalMode, string> = {
  create: "Create",
  edit: "Edit",
  delete: "Delete",
}

