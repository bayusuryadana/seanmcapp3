package service

import (
	"seanmcapp/external"
	"seanmcapp/repository"
)

// ---- WalletRepo fake ----

type fakeWalletRepo struct {
	getAllFn         func() ([]repository.Wallet, error)
	getAllocationsFn func() (map[string]int, error)
	insertFn         func(repository.Wallet) (int, error)
	updateFn         func(repository.Wallet) (int, error)
	deleteFn         func(int) (int, error)
}

func (f *fakeWalletRepo) GetAll() ([]repository.Wallet, error)    { return f.getAllFn() }
func (f *fakeWalletRepo) GetAllocations() (map[string]int, error) { return f.getAllocationsFn() }
func (f *fakeWalletRepo) Insert(w repository.Wallet) (int, error) { return f.insertFn(w) }
func (f *fakeWalletRepo) Update(w repository.Wallet) (int, error) { return f.updateFn(w) }
func (f *fakeWalletRepo) Delete(id int) (int, error)              { return f.deleteFn(id) }

// ---- StockRepo fake ----

type fakeStockRepo struct {
	getAllFn func() ([]repository.Stock, error)
	createFn func(repository.Stock) (string, error)
	updateFn func(repository.Stock) (string, error)
	deleteFn func(string) (string, error)

	updated []repository.Stock // records Update calls
}

func (f *fakeStockRepo) GetAll() ([]repository.Stock, error) { return f.getAllFn() }
func (f *fakeStockRepo) Create(s repository.Stock) (string, error) {
	return f.createFn(s)
}
func (f *fakeStockRepo) Update(s repository.Stock) (string, error) {
	f.updated = append(f.updated, s)
	if f.updateFn != nil {
		return f.updateFn(s)
	}
	return s.Name, nil
}
func (f *fakeStockRepo) Delete(name string) (string, error) { return f.deleteFn(name) }

// ---- InstagramAccountRepo fake ----

type fakeInstagramRepo struct {
	getAllFn func() ([]repository.InstagramAccount, error)
	updateFn func(username, shortcodes string) error

	updatedShortcodes map[string]string
	updatedUserIDs    map[string]string
}

func (f *fakeInstagramRepo) GetAll() ([]repository.InstagramAccount, error) { return f.getAllFn() }
func (f *fakeInstagramRepo) UpdateLastShortcodes(username, shortcodes string) error {
	if f.updatedShortcodes == nil {
		f.updatedShortcodes = map[string]string{}
	}
	f.updatedShortcodes[username] = shortcodes
	if f.updateFn != nil {
		return f.updateFn(username, shortcodes)
	}
	return nil
}

func (f *fakeInstagramRepo) UpdateUserID(username, userID string) error {
	if f.updatedUserIDs == nil {
		f.updatedUserIDs = map[string]string{}
	}
	f.updatedUserIDs[username] = userID
	return nil
}

// ---- StockClient fake ----

type fakeStockClient struct {
	prices map[string]int64
	err    error
	calls  []string
}

func (f *fakeStockClient) GetPrice(name string) (int64, error) {
	f.calls = append(f.calls, name)
	if f.err != nil {
		return 0, f.err
	}
	return f.prices[name], nil
}

// ---- TelegramClient fake ----

type telegramMessage struct {
	chatID int64
	text   string
}

type telegramPhoto struct {
	chatID  int64
	url     string
	caption string
}

type fakeTelegramClient struct {
	messages []telegramMessage
	photos   []telegramPhoto
	err      error
}

func (f *fakeTelegramClient) SendMessage(chatID int64, text string) (external.TelegramResponse, error) {
	f.messages = append(f.messages, telegramMessage{chatID, text})
	return external.TelegramResponse{Ok: true}, f.err
}

func (f *fakeTelegramClient) SendPhoto(chatID int64, photoURL, caption string) (external.TelegramResponse, error) {
	f.photos = append(f.photos, telegramPhoto{chatID, photoURL, caption})
	return external.TelegramResponse{Ok: true}, f.err
}

// ---- InstagramClient fake ----

type fakeInstagramClient struct {
	getFn func(url string) ([]byte, error)
}

func (f *fakeInstagramClient) Get(url string) ([]byte, error) { return f.getFn(url) }
