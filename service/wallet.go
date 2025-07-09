package service

import (
	"seanmcapp/repository"
	"sort"
)

type WalletService interface {
	Dashboard(date int) (*DashboardView, error)
	Create(wallet DashboardWallet) (int, error)
	Update(wallet DashboardWallet) (int, error)
	Delete(id int) (int, error)
}

type WalletServiceImpl struct {
	WalletRepo repository.WalletRepo
}

var expenseSet = map[string]struct{}{
	"Daily": {}, "Rent": {}, "Zakat": {}, "Travel": {}, "Fashion": {},
	"IT Stuff": {}, "Misc": {}, "Wellness": {}, "Funding": {},
}

func (s *WalletServiceImpl) Dashboard(date int) (*DashboardView, error) {
	wallets, err := s.WalletRepo.GetAll()
	if err != nil {
		return nil, err
	}

	dashboardBalance := calculateBalance(wallets, date)
	year := date / 100
	lastYearExpenses := calculateCategoryAmount(wallets, year-1)
	ytdExpenses := calculateCategoryAmount(wallets, year)

	currentDBS := calculateTotalAmount(wallets, "DBS", nil)
	currentBCA := calculateTotalAmount(wallets, "BCA", nil)
	plannedSGD := calculateTotalAmount(wallets, "DBS", &date)
	plannedIDR := calculateTotalAmount(wallets, "BCA", &date)

	var dashboardWallets []DashboardWallet
	for _, w := range wallets {
		if w.Date == date {
			dashboardWallets = append(dashboardWallets, DashboardWallet(w))
		}
	}

	return &DashboardView{
		Chart: DashboardChart{
			BalanceHistory:   dashboardBalance,
			LastYearExpenses: lastYearExpenses,
			YTDExpenses:      ytdExpenses,
		},
		Savings: DashboardSavings{DBS: currentDBS, BCA: currentBCA},
		Planned: DashboardPlanned{SGD: plannedSGD, IDR: plannedIDR},
		Wallets: dashboardWallets,
	}, nil
}

func calculateBalance(wallets []repository.Wallet, upToDate int) []DashboardBalance {
	balanceMap := make(map[int]int)
	for _, w := range wallets {
		if w.Account == "DBS" && w.Date <= upToDate {
			balanceMap[w.Date] += w.Amount
		}
	}
	var balances []DashboardBalance
	for date, sum := range balanceMap {
		balances = append(balances, DashboardBalance{Date: date, Sum: sum})
	}
	sort.Slice(balances, func(i, j int) bool { return balances[i].Date < balances[j].Date })

	var cumulative []DashboardBalance
	total := 0
	for _, b := range balances {
		total += b.Sum
		cumulative = append(cumulative, DashboardBalance{Date: b.Date, Sum: total})
	}
	if len(cumulative) > 12 {
		cumulative = cumulative[len(cumulative)-12:]
	}
	sort.Slice(cumulative, func(i, j int) bool { return cumulative[i].Date > cumulative[j].Date })
	return cumulative
}

func calculateCategoryAmount(wallets []repository.Wallet, year int) map[string]int {
	result := make(map[string]int)
	for _, w := range wallets {
		if w.Account == "DBS" && w.Done && (w.Date/100) == year {
			if _, ok := expenseSet[w.Category]; ok {
				result[w.Category] -= w.Amount
			}
		}
	}
	return result
}

func calculateTotalAmount(wallets []repository.Wallet, account string, date *int) int {
	total := 0
	for _, w := range wallets {
		match := w.Account == account
		if date != nil {
			match = match && (*date >= w.Date)
		} else {
			match = match && w.Done
		}
		if match {
			total += w.Amount
		}
	}
	return total
}

func (s *WalletServiceImpl) Create(wallet DashboardWallet) (int, error) {
	w := repository.Wallet(wallet)
	return s.WalletRepo.Insert(w)
}

func (s *WalletServiceImpl) Update(wallet DashboardWallet) (int, error) {
	w := repository.Wallet(wallet)
	return s.WalletRepo.Update(w)
}

func (s *WalletServiceImpl) Delete(id int) (int, error) {
	return s.WalletRepo.Delete(id)
}

type DashboardView struct {
	Chart   DashboardChart    `json:"chart"`
	Savings DashboardSavings  `json:"savings"`
	Planned DashboardPlanned  `json:"planned"`
	Wallets []DashboardWallet `json:"detail"`
}

type DashboardChart struct {
	BalanceHistory   []DashboardBalance `json:"balance"`
	LastYearExpenses map[string]int     `json:"last_year_expenses"`
	YTDExpenses      map[string]int     `json:"ytd_expenses"`
}

type DashboardSavings struct {
	DBS int `json:"dbs"`
	BCA int `json:"bca"`
}

type DashboardPlanned struct {
	SGD int `json:"sgd"`
	IDR int `json:"idr"`
}

type DashboardBalance struct {
	Date int `json:"date"`
	Sum  int `json:"sum"`
}

type DashboardWallet struct {
	ID       *int   `json:"id"`
	Date     int    `json:"date"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Currency string `json:"currency"`
	Amount   int    `json:"amount"`
	Done     bool   `json:"done"`
	Account  string `json:"account"`
}
